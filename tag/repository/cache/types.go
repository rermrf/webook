package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
	"webook/tag/domain"
)

var ErrKeyNotExist = redis.Nil

type TagCache interface {
	// GetAllTags 获取全局标签池缓存
	GetAllTags(ctx context.Context) ([]domain.Tag, error)
	SetAllTags(ctx context.Context, tags []domain.Tag) error
	DelAllTags(ctx context.Context) error
	// GetBizTags 获取某资源的标签缓存
	GetBizTags(ctx context.Context, biz string, bizId int64) ([]domain.Tag, error)
	SetBizTags(ctx context.Context, biz string, bizId int64, tags []domain.Tag) error
	DelBizTags(ctx context.Context, biz string, bizId int64) error
}
