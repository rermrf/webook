package ioc

import (
	"webook/pkg/logger"
	"webook/sms/service"
	"webook/sms/service/memory"
	"webook/sms/service/metrics"
	"webook/sms/service/opentelemtry"
)

func InitSMSService(l logger.LoggerV1) service.Service {
	// 这里切换验证码发送商
	return opentelemtry.NewTracingOTELService(metrics.NewPrometheusDecorator(memory.NewService(l)))
}

// 使用 限流器
//import (
//	"github.com/redis/go-redis/v9"
//	"time"
//	"webook/bff/pkg/ratelimit"
//	"webook/bff/service/sms"
//	"webook/bff/service/sms/memory"
//	limitsvc "webook/bff/service/sms/ratelimit"
//)
//
//func InitSMSService(cmd redis.Cmdable) sms.Service {
//	// 这里切换验证码发送商
//	svc := memory.NewService()
//	limiter := ratelimit.NewRedisSlidingWindowLimiter(cmd, time.Second, 3000)
//	return limitsvc.NewRateLimitSMSService(svc, limiter)
//}
