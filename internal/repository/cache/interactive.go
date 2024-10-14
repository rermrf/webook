package cache

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
	"webook/internal/domain"
)

//go:embed lua/interactive_incr_cnt.lua
var luaIncrCnt string

type InteractiveCache interface {
	IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error
	// Get 事实上，这里的 liked 和 collected 是不需要缓存的
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Set(ctx context.Context, biz string, bizId int64, intr domain.Interactive) error
}

// 方案1
// key1 => map[string]int

// 方案2
// key1_read_cnt => 10
// key1_collect_cnt => 11
// key1_like_cnt => 13

type RedisInteractiveCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

func NewRedisInteractiveCache(client redis.Cmdable) InteractiveCache {
	return &RedisInteractiveCache{
		client:     client,
		expiration: time.Hour,
	}
}

func (r *RedisInteractiveCache) IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	// 拿到的结果，可能自增成功了，可能不需要自增（key不存在）
	// 要不要返回一个 error 表达 key 不存在
	//res, err := r.client.Eval(ctx, luaIncrCnt,
	//	[]string{r.key(biz, bizId)},
	//	// read_cnt + 1
	//	"read_cnt", 1).Int()
	//if err != nil {
	//	return err
	//}
	//if res == 0 {
	//	// 一般缓存过期了
	//	return errors.New("缓存中 key 不存在")
	//}
	return r.client.Eval(ctx, luaIncrCnt,
		[]string{r.key(biz, bizId)},
		// read_cnt + 1
		"read_cnt", 1).Err()
}

func (r *RedisInteractiveCache) IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return r.client.Eval(ctx, luaIncrCnt, []string{r.key(biz, bizId)}, "like_cnt", 1).Err()
}

func (r *RedisInteractiveCache) DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return r.client.Eval(ctx, luaIncrCnt, []string{r.key(biz, bizId)}, "like_cnt", -1).Err()
}

func (r *RedisInteractiveCache) IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return r.client.Eval(ctx, luaIncrCnt, []string{r.key(biz, bizId)}, "collect_cnt", 1).Err()
}

func (r *RedisInteractiveCache) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	// 直接使用 HMGet，即便缓存中没有对应的 key，也不会返回 error
	//r.client.HMGet(ctx, r.key(biz, bizId), "read_cnt", "like_cnt", "collect_cnt")
	// 没有办法判定，缓存里面是有这个key，但是对应的 cnt 都是 0，还是说没有这个 key

	// 拿到 key 对于的值里面的所有 filed
	data, err := r.client.HGetAll(ctx, r.key(biz, bizId)).Result()
	if err != nil {
		return domain.Interactive{}, err
	}

	if len(data) == 0 {
		// 缓存不存在，系统错误，比如说，手动设置了缓存，但是忘记设置任何 fileds
		return domain.Interactive{}, ErrKeyNotExist
	}

	collectCnt, _ := strconv.ParseInt(data["collect_cnt"], 10, 64)
	likeCnt, _ := strconv.ParseInt(data["like_cnt"], 10, 64)
	readCnt, _ := strconv.ParseInt(data["read_cnt"], 10, 64)

	return domain.Interactive{
		CollectCnt: collectCnt,
		LikeCnt:    likeCnt,
		ReadCnt:    readCnt,
	}, nil
}

func (r *RedisInteractiveCache) Set(ctx context.Context, biz string, bizId int64, intr domain.Interactive) error {
	key := r.key(biz, bizId)
	err := r.client.HSet(ctx, key, "collect_cnt", intr.CollectCnt,
		"read_cnt", intr.ReadCnt,
		"like_cnt", intr.LikeCnt,
	).Err()
	if err != nil {
		return err
	}
	return r.client.Expire(ctx, key, time.Minute*15).Err()
}

func (r *RedisInteractiveCache) key(biz string, bizId int64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, bizId)
}
