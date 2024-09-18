package cache

import (
	"context"
	"fmt"
	lru "github.com/hashicorp/golang-lru/v2"
	"sync"
	"time"
)

type LocalCodeCache struct {
	cache      *lru.Cache[string, codeItem]
	lock       sync.RWMutex
	expiration time.Duration
}

func NewLocalCodeCache(c *lru.Cache[string, codeItem], expiration time.Duration) *LocalCodeCache {
	return &LocalCodeCache{
		cache:      c,
		expiration: expiration,
	}
}

func (l *LocalCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	key := l.key(biz, phone)
	now := time.Now()
	item, ok := l.cache.Get(key)
	if !ok {
		// 说明没有验证码
		l.cache.Add(key, codeItem{
			code:   code,
			cnt:    3,
			expire: now.Add(l.expiration),
		})
		return nil
	}
	// 一分钟内连续发送
	if item.expire.Sub(now) > time.Minute*9 {
		return ErrCodeSendTooMany
	}
	// 重新发送
	l.cache.Add(key, codeItem{
		code:   code,
		cnt:    3,
		expire: now.Add(l.expiration),
	})
	return nil
}

func (l *LocalCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	key := l.key(biz, phone)
	item, ok := l.cache.Get(key)
	if !ok {
		// 不存在
		return false, nil
	}
	// 超时了
	//if item.expire.{
	//}
	// 失败次数过多
	if item.cnt <= 0 {
		l.cache.Remove(key)
		return false, ErrCodeVerifyTooManyTimes
	}
	item.cnt--
	return item.code == inputCode, nil
}

func (l *LocalCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}

type codeItem struct {
	code   string
	cnt    int
	expire time.Time
}
