package repository

import (
	"context"
	"webook/tag/domain"
)

type TagRepository interface {
	CreateTag(ctx context.Context, tag domain.Tag) (int64, error)
	BindTagToBiz(ctx context.Context, biz string, bizId int64, tagIds []int64) error
	GetTags(ctx context.Context) ([]domain.Tag, error)
	GetTagById(ctx context.Context, id int64) (domain.Tag, error)
	GetTagsById(ctx context.Context, ids []int64) ([]domain.Tag, error)
	GetBizTags(ctx context.Context, biz string, bizId int64) ([]domain.Tag, error)
	GetBizIdsByTag(ctx context.Context, biz string, tagId int64, offset, limit int, sortBy string) ([]int64, error)
	CountBizByTag(ctx context.Context, biz string, tagId int64) (int64, error)
	FollowTag(ctx context.Context, uid, tagId int64) error
	UnfollowTag(ctx context.Context, uid, tagId int64) error
	CheckTagFollow(ctx context.Context, uid, tagId int64) (bool, error)
	GetUserFollowedTags(ctx context.Context, uid int64, offset, limit int) ([]domain.Tag, error)
	BatchGetBizTags(ctx context.Context, biz string, bizIds []int64) (map[int64][]domain.Tag, error)
}
