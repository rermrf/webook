package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
	"webook/tag/domain"
)

type RedisTagCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

func NewRedisTagCache(client redis.Cmdable) TagCache {
	return &RedisTagCache{
		client:     client,
		expiration: time.Hour * 24,
	}
}

func (r *RedisTagCache) GetAllTags(ctx context.Context) ([]domain.Tag, error) {
	data, err := r.client.Get(ctx, r.allTagsKey()).Bytes()
	if err != nil {
		return nil, err
	}
	var res []domain.Tag
	err = json.Unmarshal(data, &res)
	return res, err
}

func (r *RedisTagCache) SetAllTags(ctx context.Context, tags []domain.Tag) error {
	data, err := json.Marshal(tags)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.allTagsKey(), data, r.expiration).Err()
}

func (r *RedisTagCache) DelAllTags(ctx context.Context) error {
	return r.client.Del(ctx, r.allTagsKey()).Err()
}

func (r *RedisTagCache) GetBizTags(ctx context.Context, biz string, bizId int64) ([]domain.Tag, error) {
	data, err := r.client.Get(ctx, r.bizTagsKey(biz, bizId)).Bytes()
	if err != nil {
		return nil, err
	}
	var res []domain.Tag
	err = json.Unmarshal(data, &res)
	return res, err
}

func (r *RedisTagCache) SetBizTags(ctx context.Context, biz string, bizId int64, tags []domain.Tag) error {
	data, err := json.Marshal(tags)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.bizTagsKey(biz, bizId), data, r.expiration).Err()
}

func (r *RedisTagCache) DelBizTags(ctx context.Context, biz string, bizId int64) error {
	return r.client.Del(ctx, r.bizTagsKey(biz, bizId)).Err()
}

func (r *RedisTagCache) allTagsKey() string {
	return "tag:all_tags"
}

func (r *RedisTagCache) bizTagsKey(biz string, bizId int64) string {
	return fmt.Sprintf("tag:biz_tags:%s:%d", biz, bizId)
}
