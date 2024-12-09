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

func NewFeedPullEventDao(db *gorm.DB) FeedPullEventDao {
	return &feedPullEventDao{db: db}
}

func (f *feedPullEventDao) CreatePullEvent(ctx context.Context, event FeedPullEvent) error {
	return f.db.WithContext(ctx).Create(&event).Error
}

func (f *feedPullEventDao) FindPullEventList(ctx context.Context, uids []int64, timestamp, limit int64) ([]FeedPullEvent, error) {
	var events []FeedPullEvent
	err := f.db.WithContext(ctx).
		Where("uid IN ?", uids).
		Where("ctime < ?", timestamp).
		Order("ctime desc").
		Limit(int(limit)).
		Find(&events).Error
	return events, err
}

func (f *feedPullEventDao) FindPullEventListWithTyp(ctx context.Context, typ string, uids []int64, timestamp, limit int64) ([]FeedPullEvent, error) {
	var events []FeedPullEvent
	err := f.db.WithContext(ctx).
		Where("uid IN ?", uids).
		Where("ctime < ?", timestamp).
		Where("type = ?", typ).
		Order("ctime desc").
		Limit(int(limit)).
		Find(&events).Error
	return events, err
}
