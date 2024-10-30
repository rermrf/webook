package redisx

import (
	"context"
	"errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"net"
	"strconv"
	"time"
)

type PrometheusHook struct {
	vector *prometheus.SummaryVec
}

func NewPrometheusHook(opt prometheus.SummaryOpts) *PrometheusHook {
	// key_exist 是否命中缓存
	vector := prometheus.NewSummaryVec(opt, []string{"cmd", "key_exist"})
	prometheus.MustRegister(vector)
	return &PrometheusHook{vector: vector}
}

// DialHook 建立连接时的 hook
func (p *PrometheusHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		// 啥也不干
		return next(ctx, network, addr)
	}
}

// ProcessHook 执行普通命令时的 hook
func (p *PrometheusHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		startTime := time.Now()
		var err error
		defer func() {
			duration := time.Since(startTime).Milliseconds()
			//biz := ctx.Value("biz").(string)
			keyExist := errors.Is(err, redis.Nil)
			p.vector.WithLabelValues(
				cmd.Name(),
				//biz,
				strconv.FormatBool(keyExist),
			).Observe(float64(duration))
		}()
		err = next(ctx, cmd)
		return err
	}
}

func (p *PrometheusHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		return next(ctx, cmds)
	}
}

//func Use(client *redis.Client) {
//	client.AddHook()
//}
