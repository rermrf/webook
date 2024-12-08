package repository

import (
	"context"
	"encoding/json"
	"time"
	"webook/feed/domain"
	"webook/feed/repository/cache"
	"webook/feed/repository/dao"
)

var FolloweesNotFound = cache.FolloweesNotFound

type FeedEventRepository interface {
	// CreatePushEvents 批量推事件
	CreatePushEvents(ctx context.Context, events []domain.FeedEvent) error
	// CreatePullEvent 创建拉事件
	CreatePullEvent(ctx context.Context, event domain.FeedEvent) error
	// FindPullEvents 获取拉事件，也就是关注的人发件箱里面的事件
	FindPullEvents(ctx context.Context, uids []int64, timestamp, limit int64) ([]domain.FeedEvent, error)
	// FindPushEvents 获取推事件，也就是自己收件箱里面的事件
	FindPushEvents(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error)
	// FindPullEventsWithTyp 获取某个类型的拉事件
	FindPullEventsWithTyp(ctx context.Context, typ string, uids []int64, timestamp, limit int64) ([]domain.FeedEvent, error)
	// FindPushEventWithTyp 获取某个类型的推事件
	FindPushEventWithTyp(ctx context.Context, typ string, uid, timestamp, limit int64) ([]domain.FeedEvent, error)
}

type feedEventRepository struct {
	pullDao   dao.FeedPullEventDao
	pushDao   dao.FeedPushEventDao
	feedCache cache.FeedEventCache
}

func NewFeedEventRepository(pullDao dao.FeedPullEventDao, pushDao dao.FeedPushEventDao, feedCache cache.FeedEventCache) FeedEventRepository {
	return &feedEventRepository{pullDao: pullDao, pushDao: pushDao, feedCache: feedCache}
}

func (f feedEventRepository) CreatePushEvents(ctx context.Context, events []domain.FeedEvent) error {
	pushEvents := make([]dao.FeedPushEvent, 0, len(events))
	for _, event := range events {
		pushEvents = append(pushEvents, convertToPushEventDao(event))
	}
	return f.pushDao.CreatePushEvents(ctx, pushEvents)
}

func (f feedEventRepository) CreatePullEvent(ctx context.Context, event domain.FeedEvent) error {
	return f.pullDao.CreatePullEvent(ctx, convertToPullEventDao(event))
}

func (f feedEventRepository) FindPullEvents(ctx context.Context, uids []int64, timestamp, limit int64) ([]domain.FeedEvent, error) {
	events, err := f.pullDao.FindPullEventList(ctx, uids, timestamp, limit)
	if err != nil {
		return nil, err
	}
	ans := make([]domain.FeedEvent, 0, len(events))
	for _, event := range events {
		ans = append(ans, convertToPullEventDomain(event))
	}
	return ans, nil
}

func (f feedEventRepository) FindPushEvents(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error) {
	events, err := f.pushDao.GetPushEvents(ctx, uid, timestamp, limit)
	if err != nil {
		return nil, err
	}
	ans := make([]domain.FeedEvent, 0, len(events))
	for _, event := range events {
		ans = append(ans, convertToPushEventDomain(event))
	}
	return ans, nil
}

func (f feedEventRepository) FindPullEventsWithTyp(ctx context.Context, typ string, uids []int64, timestamp, limit int64) ([]domain.FeedEvent, error) {
	events, err := f.pullDao.FindPullEventListWithTyp(ctx, typ, uids, timestamp, limit)
	if err != nil {
		return nil, err
	}
	ans := make([]domain.FeedEvent, 0, len(events))
	for _, event := range events {
		ans = append(ans, convertToPullEventDomain(event))
	}
	return ans, nil
}

func (f feedEventRepository) FindPushEventWithTyp(ctx context.Context, typ string, uid, timestamp, limit int64) ([]domain.FeedEvent, error) {
	events, err := f.pushDao.GetPushEventsWithTyp(ctx, typ, uid, timestamp, limit)
	if err != nil {
		return nil, err
	}
	ans := make([]domain.FeedEvent, 0, len(events))
	for _, event := range events {
		ans = append(ans, convertToPushEventDomain(event))
	}
	return ans, nil
}

func convertToPushEventDao(event domain.FeedEvent) dao.FeedPushEvent {
	val, _ := json.Marshal(event.Ext)
	return dao.FeedPushEvent{
		Id:      event.Id,
		Uid:     event.Uid,
		Type:    string(val),
		Content: string(val),
		Ctime:   event.Ctime.Unix(),
	}
}

func convertToPullEventDao(event domain.FeedEvent) dao.FeedPullEvent {
	val, _ := json.Marshal(event.Ext)
	return dao.FeedPullEvent{
		Id:      event.Id,
		Uid:     event.Uid,
		Type:    event.Type,
		Content: string(val),
		Ctime:   event.Ctime.Unix(),
	}
}

func convertToPullEventDomain(event dao.FeedPullEvent) domain.FeedEvent {
	var ext map[string]string
	_ = json.Unmarshal([]byte(event.Content), &ext)
	return domain.FeedEvent{
		Id:    event.Id,
		Uid:   event.Uid,
		Type:  event.Type,
		Ctime: time.Unix(event.Ctime, 0),
		Ext:   ext,
	}
}

func convertToPushEventDomain(event dao.FeedPushEvent) domain.FeedEvent {
	var ext map[string]string
	_ = json.Unmarshal([]byte(event.Content), &ext)
	return domain.FeedEvent{
		Id:    event.Id,
		Uid:   event.Uid,
		Type:  event.Type,
		Ctime: time.Unix(event.Ctime, 0),
		Ext:   ext,
	}
}
