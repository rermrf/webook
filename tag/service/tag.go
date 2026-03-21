package service

import (
	"context"
	"math"
	"sort"
	"time"
	intrv1 "webook/api/proto/gen/intr/v1"
	"webook/pkg/logger"
	"webook/tag/domain"
	"webook/tag/events"
	"webook/tag/repository"
)

type tagService struct {
	repo     repository.TagRepository
	producer events.Producer
	intrSvc  intrv1.InteractiveServiceClient
	l        logger.LoggerV1
}

func NewTagService(repo repository.TagRepository, producer events.Producer, intrSvc intrv1.InteractiveServiceClient, l logger.LoggerV1) TagService {
	return &tagService{repo: repo, producer: producer, intrSvc: intrSvc, l: l}
}

func (s *tagService) CreateTag(ctx context.Context, name string, description string) (int64, error) {
	return s.repo.CreateTag(ctx, domain.Tag{
		Name:        name,
		Description: description,
	})
}

func (s *tagService) AttachTags(ctx context.Context, biz string, bizId int64, tagIds []int64) error {
	err := s.repo.BindTagToBiz(ctx, biz, bizId, tagIds)
	if err != nil {
		return err
	}
	// 异步发送同步事件到搜索服务
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		tags, err := s.repo.GetTagsById(ctx, tagIds)
		cancel()
		if err != nil {
			return
		}
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
		res := make([]string, 0, len(tags))
		for _, tag := range tags {
			res = append(res, tag.Name)
		}
		err = s.producer.ProduceSyncEvent(ctx, events.BizTags{
			Biz:   biz,
			BizId: bizId,
			Tags:  res,
		})
		cancel()
		if err != nil {
			s.l.Error("发送tag同步事件失败", logger.Error(err))
		}
	}()
	return nil
}

func (s *tagService) GetTags(ctx context.Context) ([]domain.Tag, error) {
	return s.repo.GetTags(ctx)
}

func (s *tagService) GetTagById(ctx context.Context, id int64) (domain.Tag, error) {
	return s.repo.GetTagById(ctx, id)
}

func (s *tagService) GetBizTags(ctx context.Context, biz string, bizId int64) ([]domain.Tag, error) {
	return s.repo.GetBizTags(ctx, biz, bizId)
}

func (s *tagService) GetBizIdsByTag(ctx context.Context, biz string, tagId int64, offset, limit int, sortBy string) ([]int64, error) {
	if sortBy == "newest" || sortBy == "" {
		return s.repo.GetBizIdsByTag(ctx, biz, tagId, offset, limit, "newest")
	}

	allIds, err := s.repo.GetBizIdsByTag(ctx, biz, tagId, 0, 1000, "newest")
	if err != nil {
		return nil, err
	}
	if len(allIds) == 0 {
		return []int64{}, nil
	}

	intrResp, err := s.intrSvc.GetByIds(ctx, &intrv1.GetByIdsRequest{
		Biz:    biz,
		BizIds: allIds,
	})
	if err != nil {
		s.l.Error("获取互动数据失败，降级为newest排序", logger.Error(err))
		return s.repo.GetBizIdsByTag(ctx, biz, tagId, offset, limit, "newest")
	}

	type scored struct {
		bizId int64
		score float64
	}
	items := make([]scored, 0, len(allIds))

	for _, id := range allIds {
		intr := intrResp.GetIntrs()[id]
		var sc float64
		if intr != nil {
			switch sortBy {
			case "hottest":
				sc = float64(intr.LikeCnt)*3 + float64(intr.CollectCnt)*5 + float64(intr.ReadCnt)*0.1
			case "featured":
				quality := float64(intr.LikeCnt)*3 + float64(intr.CollectCnt)*5
				readCnt := float64(intr.ReadCnt)
				if readCnt < 1 {
					readCnt = 1
				}
				sc = quality / math.Pow(readCnt+2, 0.5)
			}
		}
		items = append(items, scored{bizId: id, score: sc})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].score > items[j].score
	})

	start := offset
	if start > len(items) {
		return []int64{}, nil
	}
	end := start + limit
	if end > len(items) {
		end = len(items)
	}
	result := make([]int64, 0, end-start)
	for _, item := range items[start:end] {
		result = append(result, item.bizId)
	}
	return result, nil
}

func (s *tagService) CountBizByTag(ctx context.Context, biz string, tagId int64) (int64, error) {
	return s.repo.CountBizByTag(ctx, biz, tagId)
}

func (s *tagService) FollowTag(ctx context.Context, uid, tagId int64) error {
	return s.repo.FollowTag(ctx, uid, tagId)
}

func (s *tagService) UnfollowTag(ctx context.Context, uid, tagId int64) error {
	return s.repo.UnfollowTag(ctx, uid, tagId)
}

func (s *tagService) CheckTagFollow(ctx context.Context, uid, tagId int64) (bool, error) {
	return s.repo.CheckTagFollow(ctx, uid, tagId)
}

func (s *tagService) GetUserFollowedTags(ctx context.Context, uid int64, offset, limit int) ([]domain.Tag, error) {
	return s.repo.GetUserFollowedTags(ctx, uid, offset, limit)
}

func (s *tagService) BatchGetBizTags(ctx context.Context, biz string, bizIds []int64) (map[int64][]domain.Tag, error) {
	return s.repo.BatchGetBizTags(ctx, biz, bizIds)
}
