package ioc

import (
	"github.com/robfig/cron/v3"
	"time"
	"webook/internal/job"
	"webook/internal/pkg/logger"
	"webook/internal/service"
)

func InitRankingJob(svc service.RankingService) *job.RankingJob {
	return job.NewRankingJob(svc, time.Second*30)
}

func InitJob(l logger.LoggerV1, ranking *job.RankingJob) *cron.Cron {
	res := cron.New(cron.WithSeconds())
	cdb := job.NewCronJobBuilder(l)
	// 每三分钟一次
	_, err := res.AddJob("0 */3 * * * ?", cdb.Build(ranking))
	if err != nil {
		panic(err)
	}
	return res
}
