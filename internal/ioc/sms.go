package ioc

import (
	"webook/internal/service/sms"
	"webook/internal/service/sms/memory"
)

func InitSMSService() sms.Service {
	// 这里切换验证码发送商
	return memory.NewService()
}

// 使用 限流器
//import (
//	"github.com/redis/go-redis/v9"
//	"time"
//	"webook/internal/pkg/ratelimit"
//	"webook/internal/service/sms"
//	"webook/internal/service/sms/memory"
//	limitsvc "webook/internal/service/sms/ratelimit"
//)
//
//func InitSMSService(cmd redis.Cmdable) sms.Service {
//	// 这里切换验证码发送商
//	svc := memory.NewService()
//	limiter := ratelimit.NewRedisSlidingWindowLimiter(cmd, time.Second, 3000)
//	return limitsvc.NewRateLimitSMSService(svc, limiter)
//}
