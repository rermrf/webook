package service

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/sync/errgroup"
	"sort"
	"sync"
	"time"
	followv1 "webook/api/proto/gen/follow/v1"
	"webook/feed/domain"
	"webook/feed/repository"
)

type ArticleEventHandler struct {
	repo         repository.FeedEventRepository
	followClient followv1.FollowServiceClient
}

func NewArticleEventHandler(repo repository.FeedEventRepository, followClient followv1.FollowServiceClient) *ArticleEventHandler {
	return &ArticleEventHandler{repo: repo, followClient: followClient}
}

const (
	ArticleEventName = "article_event"
	// 可以调大或者调小
	// 调大，数据量大，但是用户体验好
	// 调小，数据量小，但是用户体验差
	threshold = 4
)

func (a *ArticleEventHandler) CreateFeedEvent(ctx context.Context, ext domain.ExtendFields) error {
	// 要灵活的判定是拉模型（读扩散）还是推模型（写扩散）
	uid, err := ext.Get("uid").AsInt64()
	if err != nil {
		return err
	}
	// 根据粉丝数判断使用推模型还是拉模型
	resp, err := a.followClient.GetFollowStatic(ctx, &followv1.GetFollowStaticRequest{
		Followee: uid,
	})
	if err != nil {
		return err
	}
	// 粉丝数超出阈值使用拉模型（读扩散）
	if resp.FollowStatic.Followers > threshold {
		return a.repo.CreatePullEvent(ctx, domain.FeedEvent{
			Uid:   uid,
			Type:  ArticleEventName,
			Ctime: time.Now(),
			Ext:   ext,
		})
	} else {
		// 使用推模型（写扩散）
		// 获取粉丝
		fresp, err := a.followClient.GetFollower(ctx, &followv1.GetFollowerRequest{
			Followee: uid,
		})
		if err != nil {
			return err
		}
		// 在这里，判定写扩散还是读扩散
		// 要综合考虑活跃用户，是不是铁粉
		// 在这里判定，前提是没有上面那个分支
		events := make([]domain.FeedEvent, 0, len(fresp.FollowRelations))
		for _, relation := range fresp.FollowRelations {
			events = append(events, domain.FeedEvent{
				Uid:   relation.Follower,
				Type:  ArticleEventName,
				Ctime: time.Now(),
				Ext:   ext,
			})
		}
		return a.repo.CreatePushEvents(ctx, events)
	}
}

func (a *ArticleEventHandler) FindFeedEvents(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error) {
	// 获取推模型事件
	var (
		eg errgroup.Group
		mu sync.Mutex
	)
	events := make([]domain.FeedEvent, 0, limit*2)
	// Push Event
	eg.Go(func() error {
		pushEvents, err := a.repo.FindPushEventWithTyp(ctx, ArticleEventName, uid, timestamp, limit)
		if err != nil {
			return err
		}
		mu.Lock()
		events = append(events, pushEvents...)
		mu.Unlock()
		return nil
	})

	// Pull Event
	eg.Go(func() error {
		resp, rerr := a.followClient.GetFollowee(ctx, &followv1.GetFolloweeRequest{
			Follower: uid,
			Offset:   0,
			Limit:    200,
		})
		if rerr != nil {
			return rerr
		}
		followeeIds := make([]int64, 0, len(resp.FollowRelations))
		for _, relation := range resp.FollowRelations {
			followeeIds = append(followeeIds, relation.Followee)
		}
		pullEvents, err := a.repo.FindPullEventsWithTyp(ctx, ArticleEventName, followeeIds, timestamp, limit)
		if err != nil {
			return err
		}
		mu.Lock()
		events = append(events, pullEvents...)
		mu.Unlock()
		return nil
	})
	err := eg.Wait()
	if err != nil {
		return nil, err
	}
	// 获取拉模型事件
	// 获取默认的关注列表
	sort.Slice(events, func(i, j int) bool {
		return events[i].Ctime.Unix() > events[j].Ctime.Unix()
	})

	return events[:slice.Min[int]([]int{int(limit), len(events)})], nil
}
