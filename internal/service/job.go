package service

import (
	"context"
	"time"
	"webook/internal/domain"
	"webook/internal/repository"
	logger2 "webook/pkg/logger"
)

// JobService 抢占式任务调度
type JobService interface {
	// Preempt 抢占
	Preempt(ctx context.Context) (domain.Job, error)
	ResetNextTime(ctx context.Context, j domain.Job) error
	// PreemptV1 直接返回一个释放的方法，然后调用者去调
	// PreemptV1(ctx context.Context) (domain.Job, error)
	// Release 释放
	//Release(ctx context.Context, id int64) error
}

type cronJobService struct {
	repo            repository.JobRepository
	refreshInterval time.Duration
	l               logger2.LoggerV1
}

func (p *cronJobService) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := p.repo.Preempt(ctx)

	// 续约
	//ch := make(chan struct{})
	//go func() {
	//	// 在这里续约
	//	ticker := time.NewTicker(p.refreshInterval)
	//	for {
	//		select {
	//		case <-ticker.C:
	//			p.refresh(j.Id)
	//		case <-ch:
	//			// 结束
	//			return
	//		}
	//	}
	//}()

	ticker := time.NewTicker(p.refreshInterval)
	go func() {
		for range ticker.C {
			p.refresh(j.Id)
		}
	}()

	// 抢占之后，一直抢占着吗？
	j.CancelFunc = func() error {
		//close(ch)
		// 在这里释放掉
		ticker.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		return p.repo.Release(ctx, j.Id)
	}
	return j, err
}

func (p *cronJobService) refresh(id int64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 如何续约？
	// 更新一下更新时间就可以了
	// 比如说我们的续约失败逻辑就是：处于 running 状态，但是更新时间在三分钟之前，就说明没有续约
	err := p.repo.UpdateUtime(ctx, id)
	if err != nil {
		// 可以考虑重试
		p.l.Error("续约失败", logger2.Error(err), logger2.Int64("jid", id))
	}
}

func (p *cronJobService) ResetNextTime(ctx context.Context, j domain.Job) error {
	next := j.NextTime()
	if next.IsZero() {
		// 没有下一次
		return p.repo.Stop(ctx, j.Id)
	}
	return p.repo.UpdateNextTime(ctx, j.Id, next)
}
