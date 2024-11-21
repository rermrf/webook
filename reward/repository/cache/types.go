package cache

import (
	"context"
	"webook/reward/domain"
)

type RewardCache interface {
	GetCachedCodeURL(ctx context.Context, r domain.Reward) (domain.CodeURL, error)
	CacheCodeURL(ctx context.Context, cu domain.CodeURL, r domain.Reward) error
}
