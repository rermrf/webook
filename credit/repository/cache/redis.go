package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrCacheMiss = errors.New("cache miss")

type CreditRedisCache struct {
	client redis.Cmdable
}

func NewCreditRedisCache(client redis.Cmdable) CreditCache {
	return &CreditRedisCache{client: client}
}

func (c *CreditRedisCache) balanceKey(uid int64) string {
	return fmt.Sprintf("credit:balance:%d", uid)
}

// GetBalance 获取缓存的余额
func (c *CreditRedisCache) GetBalance(ctx context.Context, uid int64) (int64, error) {
	key := c.balanceKey(uid)
	balance, err := c.client.Get(ctx, key).Int64()
	if errors.Is(err, redis.Nil) {
		return 0, ErrCacheMiss
	}
	return balance, err
}

// SetBalance 缓存余额，10分钟过期
func (c *CreditRedisCache) SetBalance(ctx context.Context, uid int64, balance int64) error {
	key := c.balanceKey(uid)
	return c.client.Set(ctx, key, balance, 10*time.Minute).Err()
}

// DelBalance 删除余额缓存
func (c *CreditRedisCache) DelBalance(ctx context.Context, uid int64) error {
	key := c.balanceKey(uid)
	return c.client.Del(ctx, key).Err()
}
