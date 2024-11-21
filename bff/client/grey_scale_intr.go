package client

import (
	"context"
	"go.uber.org/atomic"
	"google.golang.org/grpc"
	"log"
	"math/rand/v2"
	intrv1 "webook/api/proto/gen/intr/v1"
)

type GreyScaleInteractiveServiceClient struct {
	remote intrv1.InteractiveServiceClient
	local  intrv1.InteractiveServiceClient
	// 怎么去控制流量
	// 如果一个请求过来，该怎么控制它去调用本地，还是调用远程？
	// 用随机数 + 阈值的方式
	threshold atomic.Int32
}

func NewGreyScaleInteractiveServiceClient(remote intrv1.InteractiveServiceClient, local intrv1.InteractiveServiceClient) *GreyScaleInteractiveServiceClient {
	return &GreyScaleInteractiveServiceClient{
		remote:    remote,
		local:     local,
		threshold: atomic.Int32{},
	}
}

func (g *GreyScaleInteractiveServiceClient) IncrReadCnt(ctx context.Context, in *intrv1.IncrReadCntRequest, opts ...grpc.CallOption) (*intrv1.IncrReadCntResponse, error) {
	return g.client().IncrReadCnt(ctx, in, opts...)
}

func (g *GreyScaleInteractiveServiceClient) Like(ctx context.Context, in *intrv1.LikeRequest, opts ...grpc.CallOption) (*intrv1.LikeResponse, error) {
	return g.client().Like(ctx, in, opts...)
}

func (g *GreyScaleInteractiveServiceClient) CancelLike(ctx context.Context, in *intrv1.CancelLikeRequest, opts ...grpc.CallOption) (*intrv1.CancelLikeResponse, error) {
	return g.client().CancelLike(ctx, in, opts...)
}

func (g *GreyScaleInteractiveServiceClient) Collect(ctx context.Context, in *intrv1.CollectRequest, opts ...grpc.CallOption) (*intrv1.CollectResponse, error) {
	return g.client().Collect(ctx, in, opts...)
}

func (g *GreyScaleInteractiveServiceClient) CancelCollect(ctx context.Context, in *intrv1.CancelCollectRequest, opts ...grpc.CallOption) (*intrv1.CancelCollectResponse, error) {
	return g.client().CancelCollect(ctx, in, opts...)
}

func (g *GreyScaleInteractiveServiceClient) Get(ctx context.Context, in *intrv1.GetRequest, opts ...grpc.CallOption) (*intrv1.GetResponse, error) {
	return g.client().Get(ctx, in, opts...)
}

func (g *GreyScaleInteractiveServiceClient) GetByIds(ctx context.Context, in *intrv1.GetByIdsRequest, opts ...grpc.CallOption) (*intrv1.GetByIdsResponse, error) {
	return g.client().GetByIds(ctx, in, opts...)
}

func (g *GreyScaleInteractiveServiceClient) UpdateThreshold(newThreshold int32) {
	g.threshold.Store(newThreshold)
}

func (g *GreyScaleInteractiveServiceClient) client() intrv1.InteractiveServiceClient {
	threshold := g.threshold.Load()
	// 生成一个 0-100 的随机数
	num := rand.Int32N(100)
	// 如果 threshold 是100，则所有请求都是远程调用
	log.Println(num, "-----", g.threshold)
	if num <= threshold {
		return g.remote
	}
	// 如果 threshold 为0，则所有请求走本地
	return g.local
}
