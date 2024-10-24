package job

import (
	"context"
	"time"
	"webook/internal/service"
)

type RankingJob struct {
	svc     service.RankingService
	timeout time.Duration
}

func NewRankingJob(svc service.RankingService, timeout time.Duration) *RankingJob {
	return &RankingJob{
		svc: svc,
		// 根据你的数据量来，如果要是7天内的数量很多，就要设置长点
		timeout: timeout,
	}
}

func (r RankingJob) Name() string {
	return "ranking"
}

func (r RankingJob) Run(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	return r.svc.TopN(ctx)
}
