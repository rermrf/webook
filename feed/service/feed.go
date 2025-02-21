package service

import (
	"context"
	"fmt"
	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/sync/errgroup"
	"sort"
	"sync"
	followv1 "webook/api/proto/gen/follow/v1"
	"webook/feed/domain"
	"webook/feed/repository"
)

type feedService struct {
	repo repository.FeedEventRepository
	// key 就是 type，value 就是具体的业务处理逻辑
	handlerMap   map[string]Handler
	followClient followv1.FollowServiceClient
}

func NewFeedService(repo repository.FeedEventRepository, handlerMap map[string]Handler, followClient followv1.FollowServiceClient) FeedService {
	return &feedService{repo: repo, handlerMap: handlerMap, followClient: followClient}
}

func (f *feedService) registerService(typ string, handler Handler) {
	f.handlerMap[typ] = handler
}

// CreateFeedEvent 在service中根据type调用不同的handler
func (f *feedService) CreateFeedEvent(ctx context.Context, feed domain.FeedEvent) error {
	// 需要可以解决的handler
	handler, ok := f.handlerMap[feed.Type]
	if !ok {
		// 这里可以考虑引入一个兜底的处理机制
		// 例如在找不到的时候就默认丢过去 PushEvent 里面
		// 对于大部分业务来说，都是合适的
		return fmt.Errorf("未找到具体的业务处理逻辑 %s", feed.Type)
	}
	return handler.CreateFeedEvent(ctx, feed.Ext)
}

func (f *feedService) GetFeedEventList(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error) {
	// 万一，有一些业务有自己的查询逻辑；另外一些业务没有特殊的查询逻辑
	// 怎么写代码？
	// 要注意尽可能减少数据库查询次数，和 follow client 的调用次数
	var eg errgroup.Group
	// n 个 handler * limit
	res := make([]domain.FeedEvent, 0, limit*int64(len(f.handlerMap)))
	var mu sync.RWMutex
	for _, handler := range f.handlerMap {
		// 每一个业务方都要查询自己的，根据具体的handler实现查询收件箱或者发件箱
		// 要小心，这一步不要忘了
		h := handler
		eg.Go(func() error {
			events, err := h.FindFeedEvents(ctx, uid, timestamp, limit)
			if err != nil {
				return err
			}
			mu.Lock()
			res = append(res, events...)
			mu.Unlock()
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	// 聚合排序，难免的
	sort.Slice(res, func(i, j int) bool {
		return res[i].Ctime.Unix() > res[j].Ctime.Unix()
	})
	return res[:slice.Min[int]([]int{int(limit), len(res)})], nil
}

// GetFeedEventListV1 不依赖 Handler 的直接查询
// Service 层面上的统一实现
// 基本思路就是，收件箱查一下，发件箱查一下，合并结果（排序，分页），返回结果。
// 按照时间戳倒叙排序
// 查询的时候，业务上不要做特殊处理
func (f *feedService) GetFeedEventListV1(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error) {
	var eg errgroup.Group
	var mu sync.RWMutex
	// 两倍的 limit，其中一半为收件箱，另外一半为发件箱
	res := make([]domain.FeedEvent, 0, limit*2)
	eg.Go(func() error {

		// 性能瓶颈大概率出现在这里
		// 可以考虑，在触发了降级的时候，或者 follow 本身触发了降级的时候，不走这个分支
		// 我怎么知道 follow 降级了呢？

		// 获得你关注所有人的 id
		resp, rerr := f.followClient.GetFollowee(ctx, &followv1.GetFolloweeRequest{
			// 你的 ID，为了获取你关注的所有人
			Follower: uid,
			Offset:   0,
			// 可以把全部全部取过来
			Limit: 100000,
		})
		if rerr != nil {
			return rerr
		}
		followeeIds := make([]int64, 0, len(resp.FollowRelations))
		for _, relation := range resp.FollowRelations {
			followeeIds = append(followeeIds, relation.Followee)
		}
		events, err := f.repo.FindPullEvents(ctx, followeeIds, timestamp, limit)
		if err != nil {
			return err
		}
		mu.Lock()
		res = append(res, events...)
		mu.Unlock()
		return nil
	})
	eg.Go(func() error {
		// 只有一次本地数据库查询，非常快
		events, err := f.repo.FindPushEvents(ctx, uid, timestamp, limit)
		if err != nil {
			return err
		}
		mu.Lock()
		res = append(res, events...)
		mu.Unlock()
		return nil
	})
	// 合并后排序
	sort.Slice(res, func(i, j int) bool {
		return res[i].Ctime.Unix() > res[j].Ctime.Unix()
	})
	err := eg.Wait()
	// 要小心不够数量。就是你想取10条。结果总共才查到了8条
	return res[:slice.Min[int]([]int{int(limit), len(res)})], err
}
