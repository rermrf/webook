package ioc

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"webook/pkg/redisx"
	"webook/user/repository/cache"
)

//func InitUserHandler(repo repository.UserRepository) service.UserService {
//	l, err := zap.NewDevelopment()
//	if err != nil {
//		panic(err)
//	}
//	return service.NewUserService(repo, l)
//}

// InitUserCache 配合 PrometheusHook 使用
func InitUserCache(client *redis.Client) cache.UserCache {
	client.AddHook(redisx.NewPrometheusHook(
		prometheus.SummaryOpts{
			Namespace: "emoji",
			Subsystem: "webook",
			Name:      "sms_resp_time",
			Help:      "统计 SMS 服务的性能数据",
			ConstLabels: map[string]string{
				"biz": "user",
			},
		}))
	return nil
}
