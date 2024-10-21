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

type UserCache interface {
	Get(ctx context.Context, id int64) (domain.User, error)
	Set(ctx context.Context, user domain.User) error
	Delete(ctx context.Context, id int64) error
}

type RedisUserCache struct {
	// 传单机 redis 可以
	// 传 cluster 的 redis 也可以
	client     redis.Cmdable
	expiration time.Duration
}

// A 用到了 B，B 一定是借口 => 这个是保证面向接口
// A 用到了 B，B 一定是 A 的字段 => 规避包变量，包方法,都非常缺乏扩展性
// A 用到了 B，A 绝对不初始化 B，而是外面注入 => 保持依赖注入和依赖反转
func NewUserCache(client redis.Cmdable) UserCache {
	return &RedisUserCache{
		client:     client,
		expiration: 15 * time.Minute,
	}
}

func (cache *RedisUserCache) Delete(ctx context.Context, id int64) error {
	key := cache.key(id)
	return cache.client.Del(ctx, key).Err()
}

// Get 只要 error 为 nil，就认为缓存里有数据
// 如果没有数据，返回一个特定的 error
func (cache *RedisUserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	//ctx = context.WithValue(ctx, "biz", "user")
	ctx = context.WithValue(ctx, "pattern", "user:info:%d")
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

func (cache *RedisUserCache) Set(ctx context.Context, user domain.User) error {
	val, err := json.Marshal(user)
	if err != nil {
		return err
	}
	key := cache.key(user.Id)
	return cache.client.Set(ctx, key, val, cache.expiration).Err()
}

func (cache *RedisUserCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}
