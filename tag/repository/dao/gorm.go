package dao

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"gorm.io/gorm"
	"time"
)

type GORMTagDao struct {
	db *gorm.DB
}

func NewGORMTagDao(db *gorm.DB) TagDao {
	return &GORMTagDao{db: db}
}

func (d *GORMTagDao) CreateTag(ctx context.Context, tag Tag) (int64, error) {
	now := time.Now().UnixMilli()
	tag.Ctime = now
	tag.Utime = now
	// 标签名已存在则复用（全局唯一）
	err := d.db.WithContext(ctx).
		Where(Tag{Name: tag.Name}).
		FirstOrCreate(&tag).Error
	return tag.Id, err
}

func (d *GORMTagDao) CreateTagBiz(ctx context.Context, tagBiz []TagBiz) error {
	if len(tagBiz) == 0 {
		return nil
	}
	now := time.Now().UnixMilli()
	for i := range tagBiz {
		tagBiz[i].Ctime = now
		tagBiz[i].Utime = now
	}
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 覆盖式的操作：先删除该 biz+biz_id 下的所有标签绑定
		first := tagBiz[0]
		err := tx.Where("biz = ? AND biz_id = ?", first.Biz, first.BizId).
			Delete(&TagBiz{}).Error
		if err != nil {
			return err
		}
		return tx.Create(&tagBiz).Error
	})
}

func (d *GORMTagDao) GetAllTags(ctx context.Context) ([]Tag, error) {
	var res []Tag
	err := d.db.WithContext(ctx).Find(&res).Error
	return res, err
}

func (d *GORMTagDao) GetTagById(ctx context.Context, id int64) (Tag, error) {
	var res Tag
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&res).Error
	return res, err
}

func (d *GORMTagDao) GetTagsByBiz(ctx context.Context, biz string, bizId int64) ([]Tag, error) {
	var tagBizs []TagBiz
	err := d.db.WithContext(ctx).Model(&TagBiz{}).
		InnerJoins("Tag").
		Where("biz = ? AND biz_id = ?", biz, bizId).Find(&tagBizs).Error
	if err != nil {
		return nil, err
	}
	return slice.Map(tagBizs, func(idx int, src TagBiz) Tag {
		return *src.Tag
	}), nil
}

func (d *GORMTagDao) GetTagsById(ctx context.Context, ids []int64) ([]Tag, error) {
	var res []Tag
	err := d.db.WithContext(ctx).Where("id IN ?", ids).Find(&res).Error
	return res, err
}

func (d *GORMTagDao) GetBizIdsByTag(ctx context.Context, biz string, tagId int64, offset, limit int, sortBy string) ([]int64, error) {
	var bizIds []int64
	query := d.db.WithContext(ctx).Model(&TagBiz{}).
		Select("biz_id").
		Where("biz = ? AND tid = ?", biz, tagId)
	switch sortBy {
	case "hottest":
		// 按热度排序需要联合查询互动数据，这里暂时用 biz_id 倒序近似
		// 真实场景应该 JOIN interactive 表按 like_cnt+read_cnt 排序
		query = query.Order("biz_id DESC")
	default:
		// newest: 默认按绑定时间倒序（最新的在前面）
		query = query.Order("ctime DESC")
	}
	err := query.Offset(offset).Limit(limit).Find(&bizIds).Error
	return bizIds, err
}

func (d *GORMTagDao) CountBizByTag(ctx context.Context, biz string, tagId int64) (int64, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&TagBiz{}).
		Where("biz = ? AND tid = ?", biz, tagId).
		Count(&count).Error
	return count, err
}

func (d *GORMTagDao) FollowTag(ctx context.Context, uid, tagId int64) error {
	now := time.Now().UnixMilli()
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Where("uid = ? AND tag_id = ?", uid, tagId).FirstOrCreate(&TagFollow{
			Uid:   uid,
			TagId: tagId,
			Ctime: now,
		})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return nil
		}
		return tx.Model(&Tag{}).Where("id = ?", tagId).
			UpdateColumn("follower_count", gorm.Expr("follower_count + 1")).Error
	})
}

func (d *GORMTagDao) UnfollowTag(ctx context.Context, uid, tagId int64) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Where("uid = ? AND tag_id = ?", uid, tagId).Delete(&TagFollow{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return nil
		}
		return tx.Model(&Tag{}).Where("id = ? AND follower_count > 0", tagId).
			UpdateColumn("follower_count", gorm.Expr("follower_count - 1")).Error
	})
}

func (d *GORMTagDao) CheckTagFollow(ctx context.Context, uid, tagId int64) (bool, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&TagFollow{}).
		Where("uid = ? AND tag_id = ?", uid, tagId).
		Count(&count).Error
	return count > 0, err
}

func (d *GORMTagDao) GetUserFollowedTags(ctx context.Context, uid int64, offset, limit int) ([]Tag, error) {
	var tags []Tag
	err := d.db.WithContext(ctx).
		Joins("JOIN tag_follows tf ON tags.id = tf.tag_id").
		Where("tf.uid = ?", uid).
		Order("tf.ctime DESC").
		Offset(offset).Limit(limit).
		Find(&tags).Error
	return tags, err
}

func (d *GORMTagDao) BatchGetTagsByBiz(ctx context.Context, biz string, bizIds []int64) (map[int64][]Tag, error) {
	var tagBizs []TagBiz
	err := d.db.WithContext(ctx).Model(&TagBiz{}).
		InnerJoins("Tag").
		Where("biz = ? AND biz_id IN ?", biz, bizIds).
		Find(&tagBizs).Error
	if err != nil {
		return nil, err
	}
	result := make(map[int64][]Tag, len(bizIds))
	for _, tb := range tagBizs {
		if tb.Tag != nil {
			result[tb.BizId] = append(result[tb.BizId], *tb.Tag)
		}
	}
	return result, nil
}
