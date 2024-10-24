package job

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"time"
	"webook/internal/pkg/logger"
)

// RankingJobAdapter 适配器适配两个不同的接口
type RankingJobAdapter struct {
	j Job
	l logger.LoggerV1
	p prometheus.Summary
}

func NewRankingJobAdapter(j Job, l logger.LoggerV1) *RankingJobAdapter {
	p := prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace: "emoji",
		Subsystem: "webook",
		Name:      "ranking_job",
		Help:      "计算热搜榜任务",
		ConstLabels: map[string]string{
			"name": j.Name(),
		},
	})
	prometheus.MustRegister(p)
	return &RankingJobAdapter{j: j, l: l, p: p}
}

func (r *RankingJobAdapter) Run() {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		r.p.Observe(float64(duration))
	}()
	err := r.j.Run(context.Background())
	if err != nil {
		r.l.Error("运行任务失败", logger.String("job", r.j.Name()), logger.Error(err))
	}
}
