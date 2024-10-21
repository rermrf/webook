package wechat

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"time"
	"webook/internal/domain"
)

// PrometheusDecorator 对微信的登录响应时间进行监控
type PrometheusDecorator struct {
	Service
	sum prometheus.Summary
}

func NewPrometheusDecorator(service Service) *PrometheusDecorator {
	sum := prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace: "emoji",
		Subsystem: "webook",
		Name:      "wechat_resp_time",
		Help:      "统计 wechat 服务的性能数据",
	})
	prometheus.MustRegister(sum)
	return &PrometheusDecorator{
		Service: service,
		sum:     sum,
	}
}

func (p *PrometheusDecorator) VerifyCode(ctx context.Context, code string, state string) (domain.WechatInfo, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		p.sum.Observe(float64(duration.Milliseconds()))
	}()
	return p.Service.VerifyCode(ctx, code, state)
}
