package cache

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"time"
	"webook/ranking/domain"
)

type RankingCache interface {
	Set(ctx context.Context, arts []domain.Article) error
	Get(ctx context.Context) ([]domain.Article, error)
}

type RankingRedisCache struct {
	client redis.Cmdable
	key    string
}

func NewRankingRedisCache(client redis.Cmdable) *RankingRedisCache {
	return &RankingRedisCache{client: client}
}

func (r *RankingRedisCache) Set(ctx context.Context, arts []domain.Article) error {
	// 可以趁机，把 article 写到缓存里面 id => article
	for i := 0; i < len(arts); i++ {
		arts[i].Content = ""
	}
	val, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	// 这个过期时间要稍微长一点，最好超过计算热榜的时间（包含重试在内）
	// 甚至可以直接永不过期（计算热榜崩溃的情况）
	return r.client.Set(ctx, r.key, val, time.Minute*10).Err()
}

func (r *RankingRedisCache) Get(ctx context.Context) ([]domain.Article, error) {
	data, err := r.client.Get(ctx, r.key).Bytes()
	if err != nil {
		return nil, err
	}
	var arts []domain.Article
	err = json.Unmarshal(data, &arts)
	return arts, err
}
