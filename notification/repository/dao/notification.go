package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// NotificationDAO 通知数据访问接口
type NotificationDAO interface {
	Insert(ctx context.Context, n Notification) (int64, error)
	BatchInsert(ctx context.Context, ns []Notification) ([]int64, error)
	FindByKeyAndChannel(ctx context.Context, key string, channel uint8) (Notification, error)
	FindByUserId(ctx context.Context, userId int64, offset, limit int) ([]Notification, error)
	FindByUserIdAndGroup(ctx context.Context, userId int64, groupType uint8, offset, limit int) ([]Notification, error)
	FindUnreadByUserId(ctx context.Context, userId int64, offset, limit int) ([]Notification, error)
	MarkAsRead(ctx context.Context, userId int64, ids []int64) error
	MarkAllAsRead(ctx context.Context, userId int64) error
	CountUnreadByGroup(ctx context.Context, userId int64) (map[uint8]int64, error)
	UpdateStatus(ctx context.Context, id int64, status uint8) error
	Delete(ctx context.Context, userId int64, id int64) error
	DeleteByUserId(ctx context.Context, userId int64) error
}

type GORMNotificationDAO struct {
	db *gorm.DB
}

func NewGORMNotificationDAO(db *gorm.DB) NotificationDAO {
	return &GORMNotificationDAO{db: db}
}

func (d *GORMNotificationDAO) Insert(ctx context.Context, n Notification) (int64, error) {
	now := time.Now().UnixMilli()
	n.Ctime = now
	n.Utime = now
	err := d.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&n).Error
	if err != nil {
		return 0, err
	}
	return n.Id, nil
}

func (d *GORMNotificationDAO) BatchInsert(ctx context.Context, ns []Notification) ([]int64, error) {
	if len(ns) == 0 {
		return nil, nil
	}
	now := time.Now().UnixMilli()
	for i := range ns {
		ns[i].Ctime = now
		ns[i].Utime = now
	}
	err := d.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		CreateInBatches(ns, 100).Error
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(ns))
	for _, n := range ns {
		ids = append(ids, n.Id)
	}
	return ids, nil
}

func (d *GORMNotificationDAO) FindByKeyAndChannel(ctx context.Context, key string, channel uint8) (Notification, error) {
	var res Notification
	err := d.db.WithContext(ctx).
		Where("key_field = ? AND channel = ?", key, channel).
		First(&res).Error
	return res, err
}

func (d *GORMNotificationDAO) FindByUserId(ctx context.Context, userId int64, offset, limit int) ([]Notification, error) {
	var res []Notification
	err := d.db.WithContext(ctx).
		Where("user_id = ?", userId).
		Order("ctime DESC").
		Offset(offset).
		Limit(limit).
		Find(&res).Error
	return res, err
}

func (d *GORMNotificationDAO) FindByUserIdAndGroup(ctx context.Context, userId int64, groupType uint8, offset, limit int) ([]Notification, error) {
	var res []Notification
	err := d.db.WithContext(ctx).
		Where("user_id = ? AND group_type = ?", userId, groupType).
		Order("ctime DESC").
		Offset(offset).
		Limit(limit).
		Find(&res).Error
	return res, err
}

func (d *GORMNotificationDAO) FindUnreadByUserId(ctx context.Context, userId int64, offset, limit int) ([]Notification, error) {
	var res []Notification
	err := d.db.WithContext(ctx).
		Where("user_id = ? AND is_read = ?", userId, false).
		Order("ctime DESC").
		Offset(offset).
		Limit(limit).
		Find(&res).Error
	return res, err
}

func (d *GORMNotificationDAO) MarkAsRead(ctx context.Context, userId int64, ids []int64) error {
	return d.db.WithContext(ctx).
		Model(&Notification{}).
		Where("user_id = ? AND id IN ?", userId, ids).
		Updates(map[string]any{
			"is_read": true,
			"utime":   time.Now().UnixMilli(),
		}).Error
}

func (d *GORMNotificationDAO) MarkAllAsRead(ctx context.Context, userId int64) error {
	return d.db.WithContext(ctx).
		Model(&Notification{}).
		Where("user_id = ? AND is_read = ?", userId, false).
		Updates(map[string]any{
			"is_read": true,
			"utime":   time.Now().UnixMilli(),
		}).Error
}

func (d *GORMNotificationDAO) CountUnreadByGroup(ctx context.Context, userId int64) (map[uint8]int64, error) {
	type Result struct {
		GroupType uint8
		Count     int64
	}
	var results []Result
	err := d.db.WithContext(ctx).
		Model(&Notification{}).
		Select("group_type, COUNT(*) as count").
		Where("user_id = ? AND is_read = ?", userId, false).
		Group("group_type").
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	res := make(map[uint8]int64, len(results))
	for _, r := range results {
		res[r.GroupType] = r.Count
	}
	return res, nil
}

func (d *GORMNotificationDAO) UpdateStatus(ctx context.Context, id int64, status uint8) error {
	return d.db.WithContext(ctx).
		Model(&Notification{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status": status,
			"utime":  time.Now().UnixMilli(),
		}).Error
}

func (d *GORMNotificationDAO) Delete(ctx context.Context, userId int64, id int64) error {
	return d.db.WithContext(ctx).
		Where("user_id = ? AND id = ?", userId, id).
		Delete(&Notification{}).Error
}

func (d *GORMNotificationDAO) DeleteByUserId(ctx context.Context, userId int64) error {
	return d.db.WithContext(ctx).
		Where("user_id = ?", userId).
		Delete(&Notification{}).Error
}

// Notification 通知表
// 索引设计：
// 1. uk_key_channel: (key_field, channel) - 唯一索引，幂等性保证
// 2. idx_receiver_channel: (receiver, ctime) - 支持按接收者查询
// 3. idx_user_ctime: (user_id, ctime DESC) - 支持按用户查询最新通知
// 4. idx_user_group: (user_id, group_type) - 支持按分组查询
// 5. idx_user_unread: (user_id, is_read) - 支持查询未读通知
type Notification struct {
	Id             int64  `gorm:"primaryKey;autoIncrement"`
	KeyField       string `gorm:"column:key_field;type:varchar(256);uniqueIndex:uk_key_channel"`
	BizId          string `gorm:"type:varchar(64)"`
	Channel        uint8  `gorm:"uniqueIndex:uk_key_channel"`
	Receiver       string `gorm:"type:varchar(256);index:idx_receiver_channel"`
	UserId         int64  `gorm:"index:idx_user_ctime;index:idx_user_group;index:idx_user_unread"`
	TemplateId     string `gorm:"type:varchar(128)"`
	TemplateParams string `gorm:"type:json"`
	Content        string `gorm:"type:text"`
	Status         uint8
	Strategy       uint8
	GroupType      uint8  `gorm:"index:idx_user_group"`
	SourceId       int64
	SourceName     string `gorm:"type:varchar(128)"`
	TargetId       int64
	TargetType     string `gorm:"type:varchar(64)"`
	TargetTitle    string `gorm:"type:varchar(256)"`
	IsRead         bool   `gorm:"index:idx_user_unread;default:false"`
	Ctime          int64  `gorm:"index:idx_user_ctime;index:idx_receiver_channel"`
	Utime          int64
}

func (Notification) TableName() string {
	return "notifications"
}
