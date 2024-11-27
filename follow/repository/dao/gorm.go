package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type GORMFollowDao struct {
	db *gorm.DB
}

func NewGORMFollowDao(db *gorm.DB) FollowDao {
	return &GORMFollowDao{db: db}
}

func (d *GORMFollowDao) FollowRelationList(ctx context.Context, follower, offset, limit int64) ([]FollowRelation, error) {
	var res []FollowRelation
	err := d.db.WithContext(ctx).
		Where("follower = ? AND status = ?", follower, FollowRelationStatusActive).
		Offset(int(offset)).
		Limit(int(limit)).
		Find(&res).Error
	return res, err
}

func (d *GORMFollowDao) FollowRelationDetail(ctx context.Context, follower int64, followee int64) (FollowRelation, error) {
	var res FollowRelation
	err := d.db.WithContext(ctx).
		Where("follower = ? AND followee = ? AND status = ?", follower, followee, FollowRelationStatusActive).
		First(&res).Error
	return res, err
}

func (d *GORMFollowDao) CreateFollowRelation(ctx context.Context, f FollowRelation) error {
	// upsert 语义
	now := time.Now().UnixMilli()
	f.Ctime = now
	f.Ctime = now
	f.Satus = FollowRelationStatusActive
	return d.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]interface{}{
			// 这代表的是关注了-取消关注-再关注
			"status": FollowRelationStatusActive,
			"utime":  now,
		}),
	}).Create(&f).Error
	// 在这里更新 FollowStatis 的计数（也是 upsert）
}

func (d *GORMFollowDao) UpdateStatus(ctx context.Context, follower int64, followee int64, status uint8) error {
	// 当前 status 就是 inactive 的呢？
	// 不需要多此一举去检测我这个数据在不在，状态对不对
	return d.db.WithContext(ctx).
		Where("follower = ? AND followee = ?", follower, followee).
		Updates(map[string]interface{}{
			"status": status,
			"utime":  time.Now().UnixMilli(),
		}).Error
}

func (d *GORMFollowDao) CntFollower(ctx context.Context, uid int64) (int64, error) {
	var cnt int64
	err := d.db.WithContext(ctx).
		Select("count(follower)").
		// 如果要是没有额外索引，不用怀疑，就是全盘扫描
		// 可以考虑在 followee 额外创建一个索引
		Where("followee = ? AND status = ?", uid, FollowRelationStatusActive).Count(&cnt).Error
	return cnt, err
}

func (d *GORMFollowDao) CntFollowee(ctx context.Context, uid int64) (int64, error) {
	var cnt int64
	err := d.db.WithContext(ctx).
		Select("count(followee)").
		// <follower, followee>
		Where("follower = ? AND status = ?", uid, FollowRelationStatusActive).Count(&cnt).Error
	return cnt, err
}
