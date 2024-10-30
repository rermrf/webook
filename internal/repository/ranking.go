package repository

import (
	"context"
	"webook/internal/domain"
	"webook/internal/repository/cache"
)

//go:generate mockgen -source=./ranking.go -package=repomocks -destination=./mocks/ranking_mock.go
type RankingRepository interface {
	ReplaceTopN(ctx context.Context, arts []domain.Article) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

func NewCachedRankingRepository(redis *cache.RankingRedisCache, local *cache.RankingLocalCache) RankingRepository {
	return &CachedRankingRepository{
		redis: redis,
		local: local,
	}
}

type CachedRankingRepository struct {
	// 使用具体实现，可读性更好，对测试不友好，因为没有面向接口编程
	redis *cache.RankingRedisCache
	local *cache.RankingLocalCache
}

func (r *CachedRankingRepository) ReplaceTopN(ctx context.Context, arts []domain.Article) error {
	// 本地缓存几乎不可能失败
	_ = r.local.Set(ctx, arts)
	return r.redis.Set(ctx, arts)
}

func (r *CachedRankingRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {
	data, err := r.local.Get(ctx)
	if err == nil {
		return data, nil
	}
	data, err = r.redis.Get(ctx)
	if err == nil {
		_ = r.local.Set(ctx, data)
	} else {
		return r.local.ForceGet(ctx)
	}
	return data, err
}
