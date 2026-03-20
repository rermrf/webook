package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// NotificationCache 通知缓存接口
type NotificationCache interface {
	// IncrUnreadCount 增加指定分组的未读数
	IncrUnreadCount(ctx context.Context, userId int64, group uint8) error
	// GetUnreadCount 获取未读数，返回按分组的map、总数、错误
	GetUnreadCount(ctx context.Context, userId int64) (map[uint8]int64, int64, error)
	// ClearUnreadCount 清空未读数缓存
	ClearUnreadCount(ctx context.Context, userId int64) error
	// SetUnreadCount 设置未读数（从数据库同步时使用）
	SetUnreadCount(ctx context.Context, userId int64, counts map[uint8]int64) error
	// PublishSSE 发布 SSE 通知（通过 Redis Pub/Sub 推送给 BFF）
	PublishSSE(ctx context.Context, userId int64, data []byte) error
}

type RedisNotificationCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

func NewRedisNotificationCache(client redis.Cmdable) NotificationCache {
	return &RedisNotificationCache{
		client:     client,
		expiration: time.Hour * 24,
	}
}

// unreadKey 未读数缓存key，使用 Hash 结构
// key: notification:unread:{userId}
// field: group number (1-interaction, 2-reply, 3-mention, ...)
// value: count
func (c *RedisNotificationCache) unreadKey(userId int64) string {
	return fmt.Sprintf("notification:unread:%d", userId)
}

func (c *RedisNotificationCache) IncrUnreadCount(ctx context.Context, userId int64, group uint8) error {
	key := c.unreadKey(userId)
	pipe := c.client.Pipeline()
	pipe.HIncrBy(ctx, key, "total", 1)
	pipe.HIncrBy(ctx, key, strconv.Itoa(int(group)), 1)
	pipe.Expire(ctx, key, c.expiration)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *RedisNotificationCache) GetUnreadCount(ctx context.Context, userId int64) (map[uint8]int64, int64, error) {
	key := c.unreadKey(userId)
	data, err := c.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, 0, err
	}
	if len(data) == 0 {
		return nil, 0, ErrKeyNotExist
	}

	var total int64
	byGroup := make(map[uint8]int64)
	for k, v := range data {
		count, _ := strconv.ParseInt(v, 10, 64)
		if count < 0 {
			count = 0
		}
		if k == "total" {
			total = count
		} else {
			groupInt, _ := strconv.Atoi(k)
			byGroup[uint8(groupInt)] = count
		}
	}
	return byGroup, total, nil
}

func (c *RedisNotificationCache) ClearUnreadCount(ctx context.Context, userId int64) error {
	key := c.unreadKey(userId)
	return c.client.Del(ctx, key).Err()
}

func (c *RedisNotificationCache) SetUnreadCount(ctx context.Context, userId int64, counts map[uint8]int64) error {
	key := c.unreadKey(userId)
	var total int64
	fields := make(map[string]any)
	for group, count := range counts {
		fields[strconv.Itoa(int(group))] = count
		total += count
	}
	fields["total"] = total

	pipe := c.client.Pipeline()
	pipe.HSet(ctx, key, fields)
	pipe.Expire(ctx, key, c.expiration)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *RedisNotificationCache) PublishSSE(ctx context.Context, userId int64, data []byte) error {
	return c.client.Publish(ctx, "notification:sse", data).Err()
}
