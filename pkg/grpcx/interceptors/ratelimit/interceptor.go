package ratelimit

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
	intrv1 "webook/api/proto/gen/intr/v1"
	"webook/pkg/logger"
	"webook/pkg/ratelimit"
)

type InterceptorBuilder struct {
	limiter       ratelimit.Limiter
	Key           string
	L             logger.LoggerV1
	ServicePrefix string

	// key 是 FullMethod，value 是默认值的 json
	//defaultValueMap     map[string]string
	//defaultValueFuncMap map[string]func() any
}

// NewInterceptorBuilder key：user-service
// 整个应用限流、整个集群限流
// "limiter:service:user:UserService" user 里面的 UserService 限流
func NewInterceptorBuilder(limiter ratelimit.Limiter, key string, l logger.LoggerV1) *InterceptorBuilder {
	return &InterceptorBuilder{limiter: limiter, Key: key, L: l}
}

func (b *InterceptorBuilder) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		limited, err := b.limiter.Limit(ctx, b.Key)
		if err != nil {
			// err 不为nil，你要考虑用保守的，还是激进的策略
			// 这是保守策略
			b.L.Error("判定限流出现问题", logger.Error(err))
			return nil, status.Error(codes.ResourceExhausted, "触发限流")
			// 激进的策略
			//return handler(ctx, req)
		}
		if limited {
			// 限流后降级返回默认值
			//defVal, ok := b.defaultValueMap[info.FullMethod]
			//if ok {
			//	err = json.Unmarshal([]byte(defVal), &resp)
			//	return defVal, err
			//}
			return nil, status.Error(codes.ResourceExhausted, "触发限流")
		}
		return handler(ctx, req)
	}
}

// BuildServerInterceptorV1 降级
func (b *InterceptorBuilder) BuildServerInterceptorV1() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		limited, err := b.limiter.Limit(ctx, b.Key)
		if err != nil || limited {
			// err 不为nil，你要考虑用保守的，还是激进的策略
			// 这是保守策略
			b.L.Error("判定限流出现问题", logger.Error(err))
			// 激进的策略
			//return handler(ctx, req)
			ctx = context.WithValue(ctx, "limited", "true")
		}
		return handler(ctx, req)
	}
}

func (b *InterceptorBuilder) BuildClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		limited, err := b.limiter.Limit(ctx, b.Key)
		if err != nil {
			// err 不为nil，你要考虑用保守的，还是激进的策略
			// 这是保守策略
			b.L.Error("判定限流出现问题", logger.Error(err))
			return status.Error(codes.ResourceExhausted, "触发限流")
			// 激进的策略
			//return handler(ctx, req)
		}
		if limited {
			return status.Error(codes.ResourceExhausted, "触发限流")
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// BuildServerInterceptorService 针对服务级别限流
func (b *InterceptorBuilder) BuildServerInterceptorService() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if strings.HasPrefix(info.FullMethod, b.ServicePrefix) {
			limited, err := b.limiter.Limit(ctx, fmt.Sprintf("limiter:service:%s", b.ServicePrefix))
			if err != nil {
				// err 不为nil，你要考虑用保守的，还是激进的策略
				// 这是保守策略
				b.L.Error("判定限流出现问题", logger.Error(err))
				return nil, status.Error(codes.ResourceExhausted, "触发限流")
				// 激进的策略
				//return handler(ctx, req)
			}
			if limited {
				return nil, status.Error(codes.ResourceExhausted, "触发限流")
			}
		}
		return handler(ctx, req)
	}
}

// BuildServerInterceptorBiz 非通用
// 业务强相关的限流不适合在这里做，应该使用装饰器模式
func (b *InterceptorBuilder) BuildServerInterceptorBiz() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if idReq, ok := req.(*intrv1.GetRequest); ok {
			limited, err := b.limiter.Limit(ctx, fmt.Sprintf("limiter:user:%s:%d", info.FullMethod, idReq.BizId))
			if err != nil {
				// err 不为nil，你要考虑用保守的，还是激进的策略
				// 这是保守策略
				b.L.Error("判定限流出现问题", logger.Error(err))
				return nil, status.Error(codes.ResourceExhausted, "触发限流")
				// 激进的策略
				//return handler(ctx, req)
			}
			if limited {
				return nil, status.Error(codes.ResourceExhausted, "触发限流")
			}
		}
		return handler(ctx, req)
	}
}
