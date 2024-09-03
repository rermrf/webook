package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
	"webook/internal/domain"
)

var ErrKeyNotExist = redis.Nil

type UserCache struct {
	// 传单机 redis 可以
	// 传 cluster 的 redis 也可以
	client     redis.Cmdable
	expiration time.Duration
}

// A 用到了 B，B 一定是借口
// A 用到了 B，B 一定是 A 的字段
// A 用到了 B，A 绝对不初始化 B，而是外面注入
func NewUserCache(client redis.Cmdable) *UserCache {
	return &UserCache{
		client:     client,
		expiration: 15 * time.Minute,
	}
}

// Get 只要 error 为 nil，就认为缓存里有数据
// 如果没有数据，返回一个特定的 error
func (cache *UserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := cache.key(id)
	// 数据不存在， err = redis.Nil
	val, err := cache.client.Get(ctx, key).Bytes()
	if err != nil {
		return domain.User{}, err
	}
	var user domain.User
	err = json.Unmarshal(val, &user)
	return user, err
}

func (cache *UserCache) Set(ctx context.Context, user domain.User) error {
	val, err := json.Marshal(user)
	if err != nil {
		return err
	}
	key := cache.key(user.Id)
	return cache.client.Set(ctx, key, val, cache.expiration).Err()
}

func (cache *UserCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}
