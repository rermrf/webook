package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BrowseHistory struct {
	Id         int64  `gorm:"primaryKey;autoIncrement"`
	UserId     int64  `gorm:"uniqueIndex:uk_user_biz;index:idx_user_utime"`
	Biz        string `gorm:"type:varchar(64);uniqueIndex:uk_user_biz"`
	BizId      int64  `gorm:"uniqueIndex:uk_user_biz"`
	BizTitle   string `gorm:"type:varchar(256)"`
	AuthorName string `gorm:"type:varchar(128)"`
	Ctime      int64
	Utime      int64 `gorm:"index:idx_user_utime"`
}

type HistoryDAO interface {
	Upsert(ctx context.Context, h BrowseHistory) error
	FindByUserId(ctx context.Context, userId int64, cursor int64, limit int) ([]BrowseHistory, error)
	DeleteByUserId(ctx context.Context, userId int64) error
}

type GORMHistoryDAO struct {
	db *gorm.DB
}

func NewGORMHistoryDAO(db *gorm.DB) HistoryDAO {
	return &GORMHistoryDAO{db: db}
}

func (d *GORMHistoryDAO) Upsert(ctx context.Context, h BrowseHistory) error {
	now := time.Now().UnixMilli()
	h.Ctime = now
	h.Utime = now
	return d.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "biz"}, {Name: "biz_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"utime":       time.Now().UnixMilli(),
			"biz_title":   h.BizTitle,
			"author_name": h.AuthorName,
		}),
	}).Create(&h).Error
}

func (d *GORMHistoryDAO) FindByUserId(ctx context.Context, userId int64, cursor int64, limit int) ([]BrowseHistory, error) {
	var res []BrowseHistory
	db := d.db.WithContext(ctx).Where("user_id = ?", userId)
	if cursor > 0 {
		db = db.Where("utime < ?", cursor)
	}
	err := db.Order("utime DESC").Limit(limit).Find(&res).Error
	return res, err
}

func (d *GORMHistoryDAO) DeleteByUserId(ctx context.Context, userId int64) error {
	return d.db.WithContext(ctx).Where("user_id = ?", userId).Delete(&BrowseHistory{}).Error
}
