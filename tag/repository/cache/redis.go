package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
	"webook/tag/domain"
)

type RedisTagCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

func NewRedisTagCache(client redis.Cmdable) TagCache {
	return &RedisTagCache{client: client}
}

func (r *RedisTagCache) GetTags(ctx context.Context, uid int64) ([]domain.Tag, error) {
	data, err := r.client.HGetAll(ctx, r.userTagsKey(uid)).Result()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, ErrKeyNotExist
	}
	res := make([]domain.Tag, 0, len(data))
	for _, v := range data {
		var t domain.Tag
		err = json.Unmarshal([]byte(v), &t)
		if err != nil {
			return nil, err
		}
		res = append(res, t)
	}
	return res, nil
}

func (r *RedisTagCache) Append(ctx context.Context, uid int64, tags ...domain.Tag) error {
	// 要放我的标签
	// list, hash, set, sorted set
	key := r.userTagsKey(uid)
	pip := r.client.Pipeline()
	for _, t := range tags {
		val, err := json.Marshal(t)
		if err != nil {
			return err
		}
		pip.HMSet(ctx, key, strconv.FormatInt(t.Id, 10), val)
	}
	// 也可以考虑永不过期
	pip.Expire(ctx, key, r.expiration)
	_, err := pip.Exec(ctx)
	return err
}

func (r *RedisTagCache) DelTags(ctx context.Context, uid int64) error {
	return r.client.Del(ctx, r.userTagsKey(uid)).Err()
}

func (r *RedisTagCache) userTagsKey(uid int64) string {
	return fmt.Sprintf("tag:user_tags:%d", uid)
}
