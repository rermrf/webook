package ioc

import (
	"context"
	"time"
	"webook/internal/domain"
	"webook/internal/job"
	"webook/internal/pkg/logger"
	"webook/internal/service"
)

func InitScheduler(l logger.LoggerV1, svc service.JobService, local *job.LocalFuncExecutor) *job.Scheduler {
	res := job.NewScheduler(l, svc)
	res.RegisterExecutor(local)
	return res
}

func InitLocalFuncExecutor(svc service.RankingService) *job.LocalFuncExecutor {
	res := job.NewLocalFuncExecutor()
	// 要在数据库里main插入一条记录
	// ranking job 的记录，通过管理任务接口来插入
	res.RegisterFuncs("ranking", func(ctx context.Context, j domain.Job) error {
		ctx, cancel := context.WithTimeout(ctx, time.Second*30)
		defer cancel()
		return svc.TopN(ctx)
	})
	return res
}