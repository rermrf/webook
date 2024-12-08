package service

import (
	"context"
	"webook/feed/domain"
)

type FeedService interface {
	CreateFeedEvent(ctx context.Context, feed domain.FeedEvent) error
	GetFeedEventList(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error)
}

type Handler interface {
	CreateFeedEvent(ctx context.Context, ext domain.ExtendFields) error
	FindFeedEvents(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error)
}
