package client

import (
	"context"
	"go.uber.org/atomic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	followv1 "webook/api/proto/gen/follow/v1"
)

// FollowClient 通过装饰器来检测Follow服务是否可用，如不可用则降级，返回nil
type FollowClient struct {
	// 这个是真实的 RPC 客户端
	followv1.FollowServiceClient

	downgrade *atomic.Bool
}

func (f *FollowClient) GetFollowee(ctx context.Context, in *followv1.GetFolloweeRequest, opts ...grpc.CallOption) (resp *followv1.GetFolloweeResponse, err error) {
	if f.downgrade.Load() {
		// 或者返回特定的error
		return nil, nil
	}
	defer func() {
		// 比如这个，限流
		if status.Code(err) == codes.Unavailable {
			f.downgrade.Store(true)
			go func() {
				// 发心跳给 follow 检测，尝试推出 downgrade 状态
			}()
		}
	}()
	resp, err = f.FollowServiceClient.GetFollowee(ctx, in)
	return
}
