package ratelimit

import (
	"context"
	"fmt"
	"webook/internal/pkg/ratelimit"
	"webook/internal/service/sms"
)

type RateLimitSMSServiceV1 struct {
	// 被装饰的接口
	// 使用组合，这是个公共字段，组合别人可以绕开装饰器
	// 可以只实现 Service 的部分方法
	sms.Service
	limiter ratelimit.Limiter
}

func NewRateLimitSMSServiceV1(svc sms.Service, limiter ratelimit.Limiter) *RateLimitSMSService {
	return &RateLimitSMSService{
		svc:     svc,
		limiter: limiter,
	}
}

func (s RateLimitSMSServiceV1) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	// 这里可以加一些代码，新特性
	limited, err := s.limiter.Limit(ctx, "sms:tencent")
	if err != nil {
		// 系统错误
		// 可以限流：保守策略，你的下游很坑的时候
		// 可以不限：你的下游很强，业务可用性要求很高，尽量容错策略
		// 包一下这个错误
		return fmt.Errorf("短信服务判断是否限流出现问题，%w", err)
	}
	if limited {
		return errLimited
	}
	err = s.Service.Send(ctx, tpl, args, numbers...)
	// 这里也可以加一些代码，新特性
	return err
}
