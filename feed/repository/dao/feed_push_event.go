package dao

import (
	"context"
	"gorm.io/gorm"
)

type FeedPushEventDao interface {
	// CreatePushEvents 创建推送事件
	CreatePushEvents(ctx context.Context, events []FeedPushEvent) error
	GetPushEvents(ctx context.Context, uid int64, timestamp, limit int64) ([]FeedPushEvent, error)
	GetPushEventsWithTyp(ctx context.Context, typ string, uid, timestamp, limit int64) ([]FeedPushEvent, error)
}

type FeedPushEvent struct {
	Id      int64  `gorm:"primaryKey,autoIncrement"`
	Uid     int64  `gorm:"column:uid;type:int(11);not null;"`
	Type    string `gorm:"column:type;type:varchar(255);comment:类型"`
	Content string `gorm:"column:content;type:text;"`
	// 发生时间
	Ctime int64 `gorm:"column:ctime;comment:发生时间"`
}

type feedPushEventDao struct {
	db *gorm.DB
}

func newFeedPushEventDao(db *gorm.DB) FeedPushEventDao {
	return &feedPushEventDao{db: db}
}

func (f feedPushEventDao) CreatePushEvents(ctx context.Context, events []FeedPushEvent) error {
	//TODO implement me
	panic("implement me")
}

func (f feedPushEventDao) GetPushEvents(ctx context.Context, uid int64, timestamp, limit int64) ([]FeedPushEvent, error) {
	//TODO implement me
	panic("implement me")
}

func (f feedPushEventDao) GetPushEventsWithTyp(ctx context.Context, typ string, uid, timestamp, limit int64) ([]FeedPushEvent, error) {
	//TODO implement me
	panic("implement me")
}
