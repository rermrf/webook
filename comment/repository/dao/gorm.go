package dao

import (
	"context"
	"gorm.io/gorm"
)

type GORMCommentDao struct {
	db *gorm.DB
}

func NewGORMCommentDAO(db *gorm.DB) CommentDao {
	return &GORMCommentDao{db: db}
}

func (d *GORMCommentDao) Insert(ctx context.Context, u Comment) error {
	return d.db.WithContext(ctx).Create(&u).Error
}

func (d *GORMCommentDao) FindByBiz(ctx context.Context, biz string, bizId, minId, limit int64) ([]Comment, error) {
	var res []Comment
	err := d.db.WithContext(ctx).Where("biz = ? AND biz_id = ? AND id < ? AND pid IS NULL", biz, bizId, minId).Limit(int(limit)).Find(&res).Error
	return res, err
}

func (d *GORMCommentDao) FindCommentList(ctx context.Context, c Comment) ([]Comment, error) {
	var res []Comment
	builder := d.db.WithContext(ctx)
	if c.Id == 0 {
		builder = builder.
			Where("biz = ?", c.Biz).
			Where("biz_id = ?", c.BizId).
			Where("root_id is null")
	} else {
		builder = builder.Where("root_id = ? or id = ?", c.Id, c.Id)
	}
	err := builder.Find(&res).Error
	return res, err
}

// FindRepliesByPid 查找评论的直接评论
func (d *GORMCommentDao) FindRepliesByPid(ctx context.Context, pid int64, offset, limit int) ([]Comment, error) {
	var res []Comment
	err := d.db.WithContext(ctx).
		Where("pid = ?", pid).
		Order("id ASC").
		Offset(offset).
		Limit(limit).
		Find(&res).Error
	return res, err
}

func (d *GORMCommentDao) Delete(ctx context.Context, c Comment) error {
	return d.db.WithContext(ctx).Delete(&Comment{Id: c.Id}).Error
}

func (d *GORMCommentDao) FindOneByIds(ctx context.Context, ids []int64) ([]Comment, error) {
	var res []Comment
	err := d.db.WithContext(ctx).
		Where("id IN (?)", ids).
		Find(&res).Error
	return res, err
}

func (d *GORMCommentDao) FindRepliesByRid(ctx context.Context, rid int64, maxId int64, limit int64) ([]Comment, error) {
	var res []Comment
	err := d.db.WithContext(ctx).
		Where("root_id = ? AND id > ?", rid, maxId).
		Order("id ASC").
		Limit(int(limit)).Find(&res).Error
	return res, err
}
