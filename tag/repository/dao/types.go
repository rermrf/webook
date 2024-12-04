package dao

import "context"

type TagDao interface {
	CreateTag(ctx context.Context, tag Tag) (int64, error)
	CreateTagBiz(ctx context.Context, tagBiz []TagBiz) error
	GetTagsByUid(ctx context.Context, uid int64) ([]Tag, error)
	GetTagsByBiz(ctx context.Context, uid int64, biz string, bizId int64) ([]Tag, error)
	GetTags(ctx context.Context, offset, limit int) ([]Tag, error)
	GetTagsById(ctx context.Context, ids []int64) ([]Tag, error)
}

type Tag struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 联合唯一索引 <uid, name>
	Name string `gorm:"type:varchar(4096)"`
	// 你有一个经典的场景，查出一个人有什么标签，所以需要在 uid 上创建一个索引
	Uid   int64 `gorm:"index"`
	Ctime int64
	Utime int64
}

type TagBiz struct {
	Id    int64  `gorm:"primaryKey,autoIncrement"`
	BizId int64  `gorm:"index:biz_type_id"`
	Biz   string `gorm:"index:biz_type_id"`
	// 冗余字段，加快查询和删除
	Uid int64 `gorm:"index"`
	Tid int64

	// TagName string
	Tag   *Tag `gorm:"foreignKey:Tid;AssociationForeignKey:Id;constraint:OnDelete:CASCADE"`
	Ctime int64
	Utime int64
}
