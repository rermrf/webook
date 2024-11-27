package service

import (
	"context"
	"time"
	"webook/pkg/logger"
	"webook/tag/domain"
	"webook/tag/events"
	"webook/tag/repository"
)

type tagService struct {
	repo     repository.TagRepository
	producer events.Producer
	l        logger.LoggerV1
}

func NewTagService(repo repository.TagRepository, producer events.Producer, l logger.LoggerV1) TagService {
	return &tagService{repo: repo, producer: producer, l: l}
}

func (s *tagService) CreateTag(ctx context.Context, uid int64, name string) (int64, error) {
	return s.repo.CreateTag(ctx, domain.Tag{
		Uid:  uid,
		Name: name,
	})
}

func (s *tagService) AttachTags(ctx context.Context, uid int64, biz string, bizId int64, tagIds []int64) error {
	err := s.repo.BindTagToBiz(ctx, uid, biz, bizId, tagIds)
	if err != nil {
		return err
	}
	// 异步发送
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
			Uid:   uid,
			Tags:  res,
		})
		cancel()
		if err != nil {
			// 记录日志
			s.l.Error("发送tag失败", logger.Error(err))
		}
	}()
	return err
}

func (s *tagService) GetTags(ctx context.Context, uid int64) ([]domain.Tag, error) {
	return s.repo.GetTags(ctx, uid)
}

func (s *tagService) GetBizTags(ctx context.Context, uid int64, biz string, bizId int64) ([]domain.Tag, error) {
	return s.repo.GetBizTags(ctx, uid, biz, bizId)
}
