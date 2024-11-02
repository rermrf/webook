package metrics

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"time"
	"webook/sms/service"
)

type PrometheusDecorator struct {
	svc    service.Service
	vector *prometheus.SummaryVec
}

func NewPrometheusDecorator(svc service.Service) *PrometheusDecorator {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "emoji",
		Subsystem: "webook",
		Name:      "sms_resp_time",
		Help:      "统计 SMS 服务的性能数据",
	}, []string{"tplId"})
	prometheus.MustRegister(vector)
	return &PrometheusDecorator{
		svc:    svc,
		vector: vector,
	}
}

func (p PrometheusDecorator) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime).Milliseconds()
		p.vector.WithLabelValues(biz).Observe(float64(duration))
	}()
	return p.svc.Send(ctx, biz, args)
}
