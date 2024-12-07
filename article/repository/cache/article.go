package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
	"webook/article/domain"
)

type ArticleCache interface {
	GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error)
	SetFirstPage(ctx context.Context, uid int64, arts []domain.Article) error
	DelFirstPage(ctx context.Context, uid int64) error

	Set(ctx context.Context, id int64, art domain.Article) error
	Get(ctx context.Context, id int64) (domain.Article, error)

	SetPub(ctx context.Context, art domain.Article) error
	GetPub(ctx context.Context, uid int64) (domain.Article, error)
	DelPub(ctx context.Context, id int64) error
}

type RedisArticleCache struct {
	cmd redis.Cmdable
}

func NewRedisArticleCache(cmd redis.Cmdable) ArticleCache {
	return &RedisArticleCache{cmd: cmd}
}

func (c *RedisArticleCache) SetPub(ctx context.Context, art domain.Article) error {
	val, err := json.Marshal(art)
	if err != nil {
		return err
	}
	return c.cmd.Set(ctx, c.pubKey(art.Id), val, time.Minute*10).Err()
}

func (c *RedisArticleCache) GetPub(ctx context.Context, uid int64) (domain.Article, error) {
	val, err := c.cmd.Get(ctx, c.pubKey(uid)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var res domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}

func (c *RedisArticleCache) DelPub(ctx context.Context, id int64) error {
	return c.cmd.Del(ctx, c.pubKey(id)).Err()
}

func (c *RedisArticleCache) Set(ctx context.Context, id int64, art domain.Article) error {
	data, err := json.Marshal(art)
	if err != nil {
		return err
	}
	// 过期时间要短，你的预期效果越不好，就越要短
	return c.cmd.Set(ctx, c.key(id), data, time.Minute).Err()
}

func (c *RedisArticleCache) Get(ctx context.Context, id int64) (domain.Article, error) {
	data, err := c.cmd.Get(ctx, c.key(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var art domain.Article
	err = json.Unmarshal(data, &art)
	return art, err
}

func (c *RedisArticleCache) GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error) {
	data, err := c.cmd.Get(ctx, c.firstPageKey(uid)).Bytes()
	if err != nil {
		return nil, err
	}
	var arts []domain.Article
	err = json.Unmarshal(data, &arts)
	return arts, err
}

func (c *RedisArticleCache) SetFirstPage(ctx context.Context, uid int64, arts []domain.Article) error {
	for i := range arts {
		// 只缓存摘要部分
		arts[i].Content = arts[i].Abstract()
	}
	data, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	return c.cmd.Set(ctx, c.firstPageKey(uid), string(data), time.Minute*10).Err()
}

func (c *RedisArticleCache) DelFirstPage(ctx context.Context, uid int64) error {
	return c.cmd.Del(ctx, c.firstPageKey(uid)).Err()
}

func (c *RedisArticleCache) pubKey(id int64) string {
	return fmt.Sprintf("article:pub:detail:%d", id)
}

func (c *RedisArticleCache) key(id int64) string {
	return fmt.Sprintf("article:%d", id)
}

func (c *RedisArticleCache) firstPageKey(uid int64) string {
	return fmt.Sprintf("article:first_page:%d", uid)
}
