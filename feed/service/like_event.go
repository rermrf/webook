package service

import (
	"context"
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

func NewLikeEventHandler(repo repository.FeedEventRepository) *LikeEventHandler {
	return &LikeEventHandler{repo: repo}
}

// CreateFeedEvent 中的 ext 里面至少需要三个 id
// liked int64：被点赞的人
// liekr int64：点赞的人
// bizId int64：被点赞的东西
// biz string
func (l *LikeEventHandler) CreateFeedEvent(ctx context.Context, ext domain.ExtendFields) error {
	liked, err := ext.Get("liked").AsInt64()
	if err != nil {
		return err
	}
	// 你考虑校验其他数据
	// 如果你用的是扩展表设计，那么这里就会调用自己业务的扩展表来存储数据
	// 如果你希望冗余存储数据，但是业务方又不愿意存，
	// 那么你在这里可以考虑回查业务获得一些数据
	return l.repo.CreatePushEvents(ctx, []domain.FeedEvent{{
		// 收件人
		Uid:   liked,
		Type:  LikeEventName,
		Ctime: time.Now(),
		Ext:   ext,
	},
	})
}

func (l *LikeEventHandler) FindFeedEvents(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error) {
	// 如果要是你在数据库存储的时候，没有冗余用户的昵称
	// BFF（业务方） 又不愿意去聚合（调用用户服务获得昵称）
	// 就得在这里查询，注入 user client 去查
	return l.repo.FindPushEventWithTyp(ctx, LikeEventName, uid, timestamp, limit)
}
