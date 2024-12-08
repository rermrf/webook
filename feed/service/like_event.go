package service

import (
	"context"
	"errors"
	"time"
	"webook/feed/domain"
	"webook/feed/repository"
)

const (
	LikeEventName = "like_event"
)

type LikeEventHandler struct {
	repo repository.FeedEventRepository
}

// CreateFeedEvent 中的 ext 里面至少需要三个 id
// liked int64：被点赞的人
// liekr int64：点赞的人
// bizId int64：被点赞的东西
func (l *LikeEventHandler) CreateFeedEvent(ctx context.Context, ext domain.ExtendFields) error {
	liked, ok := ext.Get("liked").Val.(int64)
	if !ok {
		return errors.New("liked is not int64")
	}
	// 你考虑校验其他数据
	// 如果你用的是扩展表设计，那么这里就会调用自己业务的扩展表来存储数据
	// 如果你希望冗余存储数据，但是业务方又不愿意存，
	// 那么你在这里可以考虑回查业务获得一些数据
	return l.repo.CreatePushEvents(ctx, []domain.FeedEvent{{
		Uid:   liked,
		Type:  LikeEventName,
		Ctime: time.Now(),
		Ext:   ext,
	},
	})
}

func (l *LikeEventHandler) FindFeedEvents(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error) {
	return l.repo.FindPushEventWithTyp(ctx, LikeEventName, uid, timestamp, limit)
}
