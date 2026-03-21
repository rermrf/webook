package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type IMCache interface {
	IncrUnread(ctx context.Context, userId int64, conversationId string) error
	ClearUnread(ctx context.Context, userId int64, conversationId string) error
	GetUnread(ctx context.Context, userId int64) (map[string]int64, error)
	UpdateConvScore(ctx context.Context, userId int64, conversationId string, score int64) error
	SetOnline(ctx context.Context, userId int64, instanceId string) error
	SetOffline(ctx context.Context, userId int64) error
	IsOnline(ctx context.Context, userId int64) (bool, error)
	Publish(ctx context.Context, conversationId string, data []byte) error
}

type RedisIMCache struct {
	client redis.Cmdable
}

func NewRedisIMCache(client redis.Cmdable) IMCache {
	return &RedisIMCache{client: client}
}

func (c *RedisIMCache) unreadKey(userId int64) string {
	return fmt.Sprintf("im:unread:%d", userId)
}

func (c *RedisIMCache) convKey(userId int64) string {
	return fmt.Sprintf("im:conv:%d", userId)
}

func (c *RedisIMCache) onlineKey(userId int64) string {
	return fmt.Sprintf("im:online:%d", userId)
}

func (c *RedisIMCache) msgChannel(conversationId string) string {
	return fmt.Sprintf("im:msg:%s", conversationId)
}

func (c *RedisIMCache) IncrUnread(ctx context.Context, userId int64, conversationId string) error {
	return c.client.HIncrBy(ctx, c.unreadKey(userId), conversationId, 1).Err()
}

func (c *RedisIMCache) ClearUnread(ctx context.Context, userId int64, conversationId string) error {
	return c.client.HDel(ctx, c.unreadKey(userId), conversationId).Err()
}

func (c *RedisIMCache) GetUnread(ctx context.Context, userId int64) (map[string]int64, error) {
	result, err := c.client.HGetAll(ctx, c.unreadKey(userId)).Result()
	if err != nil {
		return nil, err
	}
	unread := make(map[string]int64, len(result))
	for k, v := range result {
		count, _ := strconv.ParseInt(v, 10, 64)
		unread[k] = count
	}
	return unread, nil
}

func (c *RedisIMCache) UpdateConvScore(ctx context.Context, userId int64, conversationId string, score int64) error {
	return c.client.ZAdd(ctx, c.convKey(userId), redis.Z{
		Score:  float64(score),
		Member: conversationId,
	}).Err()
}

func (c *RedisIMCache) SetOnline(ctx context.Context, userId int64, instanceId string) error {
	return c.client.Set(ctx, c.onlineKey(userId), instanceId, 30*time.Second).Err()
}

func (c *RedisIMCache) SetOffline(ctx context.Context, userId int64) error {
	return c.client.Del(ctx, c.onlineKey(userId)).Err()
}

func (c *RedisIMCache) IsOnline(ctx context.Context, userId int64) (bool, error) {
	n, err := c.client.Exists(ctx, c.onlineKey(userId)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (c *RedisIMCache) Publish(ctx context.Context, conversationId string, data []byte) error {
	return c.client.Publish(ctx, c.msgChannel(conversationId), data).Err()
}
