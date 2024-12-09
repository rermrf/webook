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

// FeedPushEvent 写扩散，推模型，收件箱
// 这个表理论上只插入，不更新，也不删除的
// 但是可以归档
type FeedPushEvent struct {
	Id  int64 `gorm:"primaryKey,autoIncrement"`
	Uid int64 `gorm:"index;column:uid;type:int(11);not null;"`
	// Type 用来标记是什么类型的事件
	// 决定了 Content 怎么解读
	Type    string `gorm:"column:type;type:varchar(255);comment:类型"`
	Content string `gorm:"column:content;type:text;"`
	// 发生时间
	Ctime int64 `gorm:"index;column:ctime;comment:发生时间"`
	// 这个表理论上来说，是没有 Update 操作的
	//Utime int64
}

type feedPushEventDao struct {
	db *gorm.DB
}

func NewFeedPushEventDao(db *gorm.DB) FeedPushEventDao {
	return &feedPushEventDao{db: db}
}

func (f *feedPushEventDao) CreatePushEvents(ctx context.Context, events []FeedPushEvent) error {
	return f.db.WithContext(ctx).Create(&events).Error
}

func (f *feedPushEventDao) GetPushEvents(ctx context.Context, uid int64, timestamp, limit int64) ([]FeedPushEvent, error) {
	var events []FeedPushEvent
	err := f.db.WithContext(ctx).
		Where("uid = ?", uid).
		Where("ctime < ?", timestamp).
		Order("ctime desc").
		Limit(int(limit)).
		Find(&events).Error
	return events, err
}

func (f *feedPushEventDao) GetPushEventsWithTyp(ctx context.Context, typ string, uid, timestamp, limit int64) ([]FeedPushEvent, error) {
	var events []FeedPushEvent
	err := f.db.WithContext(ctx).
		Where("uid = ?", uid).
		Where("ctime < ?", timestamp).
		Where("ctime < ?", timestamp).
		Order("ctime desc").
		Limit(int(limit)).
		Find(&events).Error
	return events, err
}
