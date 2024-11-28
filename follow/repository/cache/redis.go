package cache

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"webook/follow/domain"
)

var ErrKeyNotExists = redis.Nil

type RedisFollowCache struct {
	cmd redis.Cmdable
}

const (
	// 被多少人关注
	fieldFollowerCnt = "follower_cnt"
	// 关注了多少人
	fieldFolloweeCnt = "followee_cnt"
)

func NewRedisFollowCache(cmd redis.Cmdable) FollowCache {
	return &RedisFollowCache{cmd: cmd}
}

func (r *RedisFollowCache) StaticsInfo(ctx context.Context, uid int64) (domain.FollowStatics, error) {
	data, err := r.cmd.HGetAll(ctx, r.staticsKey(uid)).Result()
	if err != nil {
		return domain.FollowStatics{}, err
	}
	if len(data) == 0 {
		return domain.FollowStatics{}, ErrKeyNotExists
	}
	var res domain.FollowStatics
	res.Followers, _ = strconv.ParseInt(data[fieldFollowerCnt], 10, 64)
	res.Followees, _ = strconv.ParseInt(data[fieldFolloweeCnt], 10, 64)
	return res, nil
}

func (r *RedisFollowCache) SetStaticsInfo(ctx context.Context, uid int64, statics domain.FollowStatics) error {
	return r.cmd.HMSet(ctx, r.staticsKey(uid), fieldFollowerCnt, statics.Followers, fieldFolloweeCnt, statics.Followees).Err()
}

func (r *RedisFollowCache) Follow(ctx context.Context, follower, followee int64) error {
	return r.updateStaticsInfo(ctx, follower, followee, 1)
}

// 更新关注人的关注数量，更新被关注者的粉丝数量
func (r *RedisFollowCache) updateStaticsInfo(ctx context.Context, follower int64, followee int64, delta int) error {
	tx := r.cmd.TxPipeline()
	// 往 tx 中增加了两个指令
	tx.HIncrBy(ctx, r.staticsKey(follower), fieldFollowerCnt, int64(delta))
	tx.HIncrBy(ctx, r.staticsKey(followee), fieldFolloweeCnt, int64(delta))
	// 执行，Exec 的时候，会把两条命令发过去 redis server 上，并且这两条命令会一起执行，中间不会有别的命令
	// redis 的事务不具备 ACID 的特性
	_, err := tx.Exec(ctx)
	return err
}

func (r *RedisFollowCache) CancelFollow(ctx context.Context, follower, followee int64) error {
	return r.updateStaticsInfo(ctx, follower, followee, -1)
}

func (r *RedisFollowCache) staticsKey(uid int64) string {
	return fmt.Sprintf("follow:statics:%d", uid)
}
