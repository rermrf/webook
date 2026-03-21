package dao

import "context"

type TagDao interface {
	CreateTag(ctx context.Context, tag Tag) (int64, error)
	CreateTagBiz(ctx context.Context, tagBiz []TagBiz) error
	GetAllTags(ctx context.Context) ([]Tag, error)
	GetTagById(ctx context.Context, id int64) (Tag, error)
	GetTagsByBiz(ctx context.Context, biz string, bizId int64) ([]Tag, error)
	GetTagsById(ctx context.Context, ids []int64) ([]Tag, error)
	GetBizIdsByTag(ctx context.Context, biz string, tagId int64, offset, limit int, sortBy string) ([]int64, error)
	CountBizByTag(ctx context.Context, biz string, tagId int64) (int64, error)
	FollowTag(ctx context.Context, uid, tagId int64) error
	UnfollowTag(ctx context.Context, uid, tagId int64) error
	CheckTagFollow(ctx context.Context, uid, tagId int64) (bool, error)
	GetUserFollowedTags(ctx context.Context, uid int64, offset, limit int) ([]Tag, error)
	BatchGetTagsByBiz(ctx context.Context, biz string, bizIds []int64) (map[int64][]Tag, error)
}

type Tag struct {
	Id          int64  `gorm:"primaryKey,autoIncrement"`
	Name        string `gorm:"type:varchar(256);uniqueIndex"`
	Description   string `gorm:"type:varchar(1024)"`
	FollowerCount int64  `gorm:"default:0"`
	Ctime         int64
	Utime         int64
}

type TagBiz struct {
	Id    int64  `gorm:"primaryKey,autoIncrement"`
	BizId int64  `gorm:"uniqueIndex:biz_type_id_tid"`
	Biz   string `gorm:"type:varchar(128);uniqueIndex:biz_type_id_tid"`
	Tid   int64  `gorm:"uniqueIndex:biz_type_id_tid"`
	Tag      *Tag   `gorm:"foreignKey:Tid;AssociationForeignKey:Id;constraint:OnDelete:CASCADE"`
	BizCtime int64  `gorm:"default:0"`
	Ctime    int64
	Utime    int64
}

type TagFollow struct {
	Id    int64 `gorm:"primaryKey,autoIncrement"`
	Uid   int64 `gorm:"uniqueIndex:uk_uid_tag"`
	TagId int64 `gorm:"uniqueIndex:uk_uid_tag;index:idx_tag_id"`
	Ctime int64
}
