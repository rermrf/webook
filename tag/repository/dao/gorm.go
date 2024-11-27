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
	err := d.db.WithContext(ctx).Create(&tag).Error
	return tag.Id, err
}

func (d *GORMTagDao) CreateTagBiz(ctx context.Context, tagBiz []TagBiz) error {
	if len(tagBiz) == 0 {
		return nil
	}
	now := time.Now().UnixMilli()
	for _, tagBiz := range tagBiz {
		tagBiz.Ctime = now
		tagBiz.Utime = now
	}
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		first := tagBiz[0]
		err := tx.Model(&TagBiz{}).
			Delete("uid = ? AND biz = ? AND biz_id = ?", first.Uid, first.BizId, first.BizId).Error
		if err != nil {
			return err
		}
		return tx.Create(&tagBiz).Error
	})
}

func (d *GORMTagDao) GetTagsByUid(ctx context.Context, uid int64) ([]Tag, error) {
	var res []Tag
	err := d.db.WithContext(ctx).Where("uid = ?", uid).Find(&res).Error
	return res, err
}

func (d *GORMTagDao) GetTagsByBiz(ctx context.Context, uid int64, biz string, bizId int64) ([]Tag, error) {
	// GORM 的 JOIN 查询
	var tagBizs []TagBiz
	err := d.db.WithContext(ctx).Model(&TagBiz{}).
		InnerJoins("Tag", d.db.Model(&Tag{})).
		// tag_bizs.uid
		Where("Tag.uid = ? AND biz = ? AND biz_id = ?", uid, biz, bizId).Find(&tagBizs).Error
	if err != nil {
		return nil, err
	}
	return slice.Map(tagBizs, func(idx int, src TagBiz) Tag {
		return *src.Tag
	}), nil
}

func (d *GORMTagDao) GetTags(ctx context.Context, offset, limit int) ([]Tag, error) {
	var res []Tag
	err := d.db.WithContext(ctx).Offset(offset).
		Limit(limit).Find(&res).Error
	return res, err
}

func (d *GORMTagDao) GetTagsById(ctx context.Context, ids []int64) ([]Tag, error) {
	var res []Tag
	err := d.db.WithContext(ctx).Where("id IN ?", ids).Find(&res).Error
	return res, err
}
