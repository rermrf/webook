package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// TemplateCache 通知模板缓存接口
type TemplateCache interface {
	Get(ctx context.Context, templateId string, channel uint8) ([]byte, error)
	Set(ctx context.Context, templateId string, channel uint8, data []byte) error
	Del(ctx context.Context, templateId string, channel uint8) error
}

type RedisTemplateCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

func NewRedisTemplateCache(client redis.Cmdable) TemplateCache {
	return &RedisTemplateCache{
		client:     client,
		expiration: time.Hour * 24,
	}
}

func (c *RedisTemplateCache) key(templateId string, channel uint8) string {
	return fmt.Sprintf("notification:template:%s:%d", templateId, channel)
}

func (c *RedisTemplateCache) Get(ctx context.Context, templateId string, channel uint8) ([]byte, error) {
	data, err := c.client.Get(ctx, c.key(templateId, channel)).Bytes()
	if err == redis.Nil {
		return nil, ErrKeyNotExist
	}
	return data, err
}

func (c *RedisTemplateCache) Set(ctx context.Context, templateId string, channel uint8, data []byte) error {
	return c.client.Set(ctx, c.key(templateId, channel), data, c.expiration).Err()
}

func (c *RedisTemplateCache) Del(ctx context.Context, templateId string, channel uint8) error {
	return c.client.Del(ctx, c.key(templateId, channel)).Err()
}
