package ioc

import (
	rlock "github.com/gotomicro/redis-lock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"webook/pkg/redisx"
)

func InitRedis() redis.Cmdable {
	type Config struct {
		Addr string `yaml:"addr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("redis", &cfg)
	if err != nil {
		panic(err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: cfg.Addr})
	rdb.AddHook(redisx.NewPrometheusHook(prometheus.SummaryOpts{
		Namespace: "emoji",
		Subsystem: "webook",
		Name:      "redis_exec_time_and_key_is_exist",
		Help:      "统计 redis 普通命令的执行时间和 key 是否命中",
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}))
	return rdb
}

func InitRLockClient(client redis.Cmdable) *rlock.Client {
	return rlock.NewClient(client)
}
