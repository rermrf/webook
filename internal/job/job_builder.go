package job

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"strconv"
	"time"
	logger2 "webook/pkg/logger"
)

// CronJobBuilder 使用summaryVec对不同的job进行监控
type CronJobBuilder struct {
	l      logger2.LoggerV1
	p      *prometheus.SummaryVec
	tracer trace.Tracer
}

func NewCronJobBuilder(l logger2.LoggerV1) *CronJobBuilder {
	opt := prometheus.SummaryOpts{
		Namespace: "emoji",
		Subsystem: "webook",
		Name:      "cron_job",
		Help:      "统计 定时任务 的执行情况",
	}
	p := prometheus.NewSummaryVec(opt, []string{"name", "success"})
	prometheus.MustRegister(p)
	return &CronJobBuilder{
		l:      l,
		p:      p,
		tracer: otel.GetTracerProvider().Tracer("webook/internal/job/job_builder.go"),
	}
}

func (b *CronJobBuilder) Build(job Job) cron.Job {
	name := job.Name()
	return cronJobFuncAdapter(func() error {
		ctx, span := b.tracer.Start(context.Background(), name)
		defer span.End()
		start := time.Now()
		b.l.Info("任务开始", logger2.String("job", name))
		var success bool
		defer func() {
			b.l.Info("任务结束", logger2.String("job", name))
			duration := time.Since(start)
			b.p.WithLabelValues(name, strconv.FormatBool(success)).Observe(float64(duration))
		}()
		err := job.Run(ctx)
		success = err == nil
		if err != nil {
			span.RecordError(err)
			b.l.Error("运行任务失败", logger2.String("job", job.Name()), logger2.Error(err))
		}
		return nil
	})
}

type cronJobFuncAdapter func() error

func (f cronJobFuncAdapter) Run() {
	_ = f()
}
