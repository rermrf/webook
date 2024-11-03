package job

import (
	"context"
	rlock "github.com/gotomicro/redis-lock"
	"sync"
	"time"
	rankingv1 "webook/api/proto/gen/ranking/v1"
	logger2 "webook/pkg/logger"
)

type RankingJob struct {
	svc       rankingv1.RankingServiceClient
	timeout   time.Duration
	client    *rlock.Client
	key       string
	l         logger2.LoggerV1
	lock      *rlock.Lock
	localLock *sync.Mutex
}

func NewRankingJob(svc rankingv1.RankingServiceClient,
	timeout time.Duration,
	client *rlock.Client,
	l logger2.LoggerV1,
) *RankingJob {
	return &RankingJob{
		svc: svc,
		// 根据你的数据量来，如果要是7天内的数量很多，就要设置长点
		timeout:   timeout,
		client:    client,
		key:       "rlock:cron_job:ranking",
		l:         l,
		localLock: &sync.Mutex{},
	}
}

func (r *RankingJob) Name() string {
	return "ranking"
}

// 按时间调度，三分钟一次
func (r *RankingJob) Run(ctx context.Context) error {
	r.localLock.Lock()
	defer r.localLock.Unlock()
	if r.lock == nil {
		// 说明你没拿到锁，得尝试拿锁

		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		lock, err := r.client.Lock(ctx, r.key, r.timeout, &rlock.FixIntervalRetry{
			Interval: time.Millisecond * 100,
			Max:      0,
		}, time.Second)
		if err != nil {
			// 这里没拿到锁，极大概率是别人持有了锁
			return nil
		}
		r.lock = lock
		// 怎么保证我这里，一直拿着这个锁
		go func() {
			// 自动续约机制
			er := lock.AutoRefresh(r.timeout/2, time.Minute)
			// 这里说明退出了续约机制
			if er != nil {
				// 不怎么办
				// 争取下一次继续抢锁
				r.l.Error("分布式锁续约失败", logger2.Error(err))
			}
			r.localLock.Lock()
			r.lock = nil
			r.localLock.Unlock()
			// lock.Unlock(ctx)
		}()
	}
	// 不需要释放锁
	//defer func() {
	//	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//	defer cancel()
	//	err := r.lock.Unlock(ctx)
	//	if err != nil {
	//		r.l.Error("释放分布式锁失败, Ranking Job", logger.Error(err))
	//	}
	//}()
	_, err := r.svc.RankTopN(ctx, &rankingv1.RankTopNRequest{})
	return err
}

func (r *RankingJob) Close() error {
	r.localLock.Lock()
	lock := r.lock
	r.lock = nil
	r.localLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return lock.Unlock(ctx)
}
