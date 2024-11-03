package cache

import (
	"context"
	"errors"
	"go.uber.org/atomic"
	"time"
	"webook/ranking/domain"
)

// 强制使用本地缓存的漏洞：
// 如果一个节点本身没有本地缓存，此时 redis 又崩溃了，那么依旧拿不到榜单数据
// 这种情况下，可以考虑走一个 fail over（容错）策略，让前端在加载不到热榜数据的情况下，重新发一个请求
// 这样一来，除非全部后端节点都没有本地缓存，redis又崩溃了，否则必然可以加载出来一个榜单数据

type RankingLocalCache struct {
	topN       *atomic.Value
	ddl        *atomic.Value
	expiration time.Duration
}

func NewRankingLocalCache() *RankingLocalCache {
	topN := &atomic.Value{}
	ddl := &atomic.Value{}
	topN.Store([]domain.Article{})
	ddl.Store(time.Now())
	return &RankingLocalCache{
		topN: topN,
		ddl:  ddl,
		// k可以永不过期，非常长的时间或者对齐到 redis 的过期时间
		expiration: time.Minute * 10,
	}
}

func (r *RankingLocalCache) Set(ctx context.Context, arts []domain.Article) error {
	// 也可以按照 id => Article 缓存
	r.topN.Store(arts)
	ddl := time.Now().Add(r.expiration)
	r.ddl.Store(ddl)
	return nil
}

func (r *RankingLocalCache) Get(ctx context.Context) ([]domain.Article, error) {
	ddl, ok := r.ddl.Load().(time.Time)
	if !ok {
		return nil, errors.New("类型错误")
	}
	arts, ok := r.topN.Load().([]domain.Article)
	if !ok {
		return nil, errors.New("类型错误")
	}
	if len(arts) == 0 || ddl.Before(time.Now()) {
		return nil, errors.New("本地缓存未命中")
	}
	return arts, nil
}

// ForceGet 兜底工作，不计算过期时间，强制返回数据
func (r *RankingLocalCache) ForceGet(ctx context.Context) ([]domain.Article, error) {
	arts, ok := r.topN.Load().([]domain.Article)
	if !ok {
		return nil, errors.New("类型错误")
	}
	return arts, nil
}

//func (r *RankingLocalCache) Preload(ctx context.Context) error {
//
//}
