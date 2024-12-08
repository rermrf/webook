package service

import (
	"context"
	"errors"
	"time"
	"webook/feed/domain"
	"webook/feed/repository"
)

const (
	FollowEventName = "follow_event"
)

type FollowEventHandler struct {
	repo repository.FeedEventRepository
}

func NewFollowEventHandler(repo repository.FeedEventRepository) *FollowEventHandler {
	return &FollowEventHandler{repo: repo}
}

// CreateFeedEvent 创建跟随方式
// 如果 A 关注了 B，那么
// follower 就是 A
// followee 就是 B
func (f *FollowEventHandler) CreateFeedEvent(ctx context.Context, ext domain.ExtendFields) error {
	followee, ok := ext.Get("followee").Val.(int64)
	if !ok {
		return errors.New("follow event not found")
	}
	return f.repo.CreatePushEvents(ctx, []domain.FeedEvent{{
		Uid:   followee,
		Type:  FollowEventName,
		Ctime: time.Now(),
		Ext:   ext,
	}})
}

func (f *FollowEventHandler) FindFeedEvents(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error) {
	return f.repo.FindPushEventWithTyp(ctx, FollowEventName, uid, timestamp, limit)
}
