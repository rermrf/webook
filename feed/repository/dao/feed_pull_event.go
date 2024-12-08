package dao

import (
	"context"
	"gorm.io/gorm"
)

// FeedPullEventDao 拉模型
type FeedPullEventDao interface {
	CreatePullEvent(ctx context.Context, event FeedPullEvent) error
	FindPullEventList(ctx context.Context, uids []int64, timestamp, limit int64) ([]FeedPullEvent, error)
	FindPullEventListWithTyp(ctx context.Context, typ string, uids []int64, timestamp, limit int64) ([]FeedPullEvent, error)
}

type FeedPullEvent struct {
	Id      int64  `gorm:"primaryKey,autoIncrement"`
	Uid     int64  `gorm:"column:uid;type:int(11);not null;"`
	Type    string `gorm:"column:type;type:varchar(255);comment:类型"`
	Content string `gorm:"column:content;type:text;"`
	// 发生时间
	Ctime int64 `gorm:"column:ctime;comment:发生时间"`
}

type feedPullEventDao struct {
	db *gorm.DB
}

func newFeedPullEventDao(db *gorm.DB) FeedPullEventDao {
	return &feedPullEventDao{db: db}
}

func (f feedPullEventDao) CreatePullEvent(ctx context.Context, event FeedPullEvent) error {
	//TODO implement me
	panic("implement me")
}

func (f feedPullEventDao) FindPullEventList(ctx context.Context, uids []int64, timestamp, limit int64) ([]FeedPullEvent, error) {
	//TODO implement me
	panic("implement me")
}

func (f feedPullEventDao) FindPullEventListWithTyp(ctx context.Context, typ string, uids []int64, timestamp, limit int64) ([]FeedPullEvent, error) {
	//TODO implement me
	panic("implement me")
}
