package circuitbreaker

import (
	"context"
	"github.com/go-kratos/aegis/circuitbreaker"
	"google.golang.org/grpc"
	rand2 "math/rand/v2"
	"time"
)

type InterceptorBuilder struct {
	breaker circuitbreaker.CircuitBreaker

	// 设置标记位
	// 假如说我们考虑使用随机数 + 阈值的回复方式
	// 触发熔断的时候，直接将 threshold 置为0
	// 后续等一段时间， 将 theshold 调整为 1，判定请求有没有问题
	threshold int
}

func (b *InterceptorBuilder) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		// Allow 是kratos 自己写的动态判定方法，自备恢复
		if b.breaker.Allow() == nil {
			resp, err = handler(ctx, req)
			//s, ok := status.FromError(err)
			//if s != nil && s.Code() == codes.Unavailable {
			//	b.breaker.MarkFailed()
			//} else {
			//
			//}
			if err != nil {
				// 没有区别业务错误和系统错误
				b.breaker.MarkFailed()
			} else {
				b.breaker.MarkSuccess()
			}
		}
		// 触发了熔断器
		b.breaker.MarkFailed()
		return nil, err
	}
}

func (b *InterceptorBuilder) BuildServerInterceptorV1() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if !b.allow() {
			// 触发了熔断
			b.threshold = 0
			time.AfterFunc(time.Minute, func() {
				b.threshold = 1
			})
		}
		// 随机数判定
		rand := rand2.IntN(100)
		if rand <= b.threshold {
			resp, err = handler(ctx, req)
			if err == nil && b.threshold != 0 {
				// 考虑调大 threshold
			} else if b.threshold != 0 {
				// 考虑调小 threshold
			}
		}
		return resp, err
	}
}

func (b *InterceptorBuilder) allow() bool {
	// 这边就套用之前在短信里面讲的，判定结点是否健康的各种做法
	// 从 prometheus 里面拿数据判定
	//prometheus.DefaultGatherer.Gather()
	return false
}
