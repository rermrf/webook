package repository

import (
	"context"
	"webook/internal/domain"
	"webook/internal/repository/cache"
)

type RankingRepository interface {
	ReplaceTopN(ctx context.Context, arts []domain.Article) error
}

func NewCachedRankingRepository(c cache.RankingCache) RankingRepository {
	return &CachedRankingRepository{c: c}
}

type CachedRankingRepository struct {
	c cache.RankingCache
}

func (r *CachedRankingRepository) ReplaceTopN(ctx context.Context, arts []domain.Article) error {
	return r.c.Set(ctx, arts)
}
