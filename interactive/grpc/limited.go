package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	intrv1 "webook/api/proto/gen/intr/v1"
	"webook/pkg/logger"
	"webook/pkg/ratelimit"
)

// LimitedInteractiveServiceServer 使用装饰器模式对业务进行限流
type LimitedInteractiveServiceServer struct {
	limiter ratelimit.Limiter
	L       logger.LoggerV1
	intrv1.InteractiveServiceServer
}

func (i *LimitedInteractiveServiceServer) IncrReadCnt(ctx context.Context, request *intrv1.IncrReadCntRequest) (*intrv1.IncrReadCntResponse, error) {
	limited, err := i.limiter.Limit(ctx, fmt.Sprintf("limiter:interavtive:incr_read_cnt:%d", request.BizId))
	if err != nil {
		// err 不为nil，你要考虑用保守的，还是激进的策略
		// 这是保守策略
		i.L.Error("判定限流出现问题", logger.Error(err))
		return nil, status.Error(codes.ResourceExhausted, "触发限流")
		// 激进的策略
		//return handler(ctx, req)
	}
	if limited {
		return nil, status.Error(codes.ResourceExhausted, "触发限流")
	}

	resp, err := i.InteractiveServiceServer.IncrReadCnt(ctx, request)
	return resp, err
}
