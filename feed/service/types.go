package service

import (
	"context"
	"webook/feed/domain"
)

// FeedService 处理业务公共的部分
// 并且找出 Handler 来处理业务的个性部分
type FeedService interface {
	CreateFeedEvent(ctx context.Context, feed domain.FeedEvent) error
	GetFeedEventList(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error)
}

// Handler 具体业务的处理逻辑
// 按照 type 来分，因为 type 是天然的标记了哪个业务
type Handler interface {
	CreateFeedEvent(ctx context.Context, ext domain.ExtendFields) error
	FindFeedEvents(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error)
}

// FeedServiceV1 不考虑分离不同业务之间的共性和个性
type FeedServiceV1 interface {
	CreateLikeFeedEvent()
	CreateArticleFeedEvent()
	CreateFollowFeedEvent()
	// 没来一个业务方，你就要在这里加一个
	// 如果你的业务很复杂，你的代码迅速腐烂
}
