package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// TransactionDAO 事务消息数据访问接口
type TransactionDAO interface {
	Insert(ctx context.Context, t NotificationTransaction) (int64, error)
	FindByKey(ctx context.Context, key string) (NotificationTransaction, error)
	UpdateStatus(ctx context.Context, key string, status uint8) error
	FindPreparedTimeout(ctx context.Context, limit int) ([]NotificationTransaction, error)
	IncrRetryCount(ctx context.Context, id int64, nextCheckTime int64) error
}

type GORMTransactionDAO struct {
	db *gorm.DB
}

func NewGORMTransactionDAO(db *gorm.DB) TransactionDAO {
	return &GORMTransactionDAO{db: db}
}

func (d *GORMTransactionDAO) Insert(ctx context.Context, t NotificationTransaction) (int64, error) {
	now := time.Now().UnixMilli()
	t.Ctime = now
	t.Utime = now
	err := d.db.WithContext(ctx).Create(&t).Error
	if err != nil {
		return 0, err
	}
	return t.Id, nil
}

func (d *GORMTransactionDAO) FindByKey(ctx context.Context, key string) (NotificationTransaction, error) {
	var res NotificationTransaction
	err := d.db.WithContext(ctx).
		Where("key_field = ?", key).
		First(&res).Error
	return res, err
}

func (d *GORMTransactionDAO) UpdateStatus(ctx context.Context, key string, status uint8) error {
	return d.db.WithContext(ctx).
		Model(&NotificationTransaction{}).
		Where("key_field = ?", key).
		Updates(map[string]any{
			"status": status,
			"utime":  time.Now().UnixMilli(),
		}).Error
}

func (d *GORMTransactionDAO) FindPreparedTimeout(ctx context.Context, limit int) ([]NotificationTransaction, error) {
	var res []NotificationTransaction
	now := time.Now().UnixMilli()
	err := d.db.WithContext(ctx).
		Where("status = ? AND next_check_time < ?", uint8(1), now).
		Order("next_check_time ASC").
		Limit(limit).
		Find(&res).Error
	return res, err
}

func (d *GORMTransactionDAO) IncrRetryCount(ctx context.Context, id int64, nextCheckTime int64) error {
	return d.db.WithContext(ctx).
		Model(&NotificationTransaction{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"retry_count":     gorm.Expr("retry_count + 1"),
			"next_check_time": nextCheckTime,
			"utime":           time.Now().UnixMilli(),
		}).Error
}

// NotificationTransaction 事务消息表
// 索引设计：
// 1. uk_notification_id: (notification_id) - 唯一索引
// 2. uk_key: (key_field) - 唯一索引，幂等性保证
// 3. idx_status_check: (status, next_check_time) - 支持超时回查
type NotificationTransaction struct {
	Id                 int64  `gorm:"primaryKey;autoIncrement"`
	NotificationId     int64  `gorm:"uniqueIndex:uk_notification_id"`
	KeyField           string `gorm:"column:key_field;type:varchar(256);uniqueIndex:uk_key"`
	BizId              string `gorm:"type:varchar(64)"`
	Status             uint8  `gorm:"index:idx_status_check"`
	CheckBackTimeoutMs int64
	NextCheckTime      int64 `gorm:"index:idx_status_check"`
	RetryCount         int
	MaxRetry           int   `gorm:"default:5"`
	Ctime              int64
	Utime              int64
}

func (NotificationTransaction) TableName() string {
	return "notification_transactions"
}
