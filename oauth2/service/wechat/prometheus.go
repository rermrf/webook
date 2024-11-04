package wechat

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"time"
	"webook/oauth2/domain"
)

// PrometheusDecorator 对微信的登录响应时间进行监控
type PrometheusDecorator struct {
	Service
	sum prometheus.Summary
}

func NewPrometheusDecorator(
	service Service,
	nameSpace string,
	subsystem string,
	name string,
	instanceId string,
) *PrometheusDecorator {
	sum := prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace: nameSpace,
		Subsystem: subsystem,
		Name:      name,
		ConstLabels: map[string]string{
			"instance_id": instanceId,
		},
		Help: "统计 wechat 服务的性能数据",
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.9:   0.01,
			0.95:  0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
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
