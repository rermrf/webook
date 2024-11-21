package client

import (
	"context"
	"google.golang.org/grpc"
	intrv1 "webook/api/proto/gen/intr/v1"
	"webook/interactive/domain"
	"webook/interactive/service"
)

// InteractiveServiceAdapter 将一个本地实现伪装成 grpc 客户端
type InteractiveServiceAdapter struct {
	svc service.InteractiveService
}

func NewInteractiveServiceAdapter(svc service.InteractiveService) *InteractiveServiceAdapter {
	return &InteractiveServiceAdapter{svc: svc}
}

func (i *InteractiveServiceAdapter) IncrReadCnt(ctx context.Context, in *intrv1.IncrReadCntRequest, opts ...grpc.CallOption) (*intrv1.IncrReadCntResponse, error) {
	err := i.svc.IncrReadCnt(ctx, in.GetBiz(), in.GetBizId())
	return &intrv1.IncrReadCntResponse{}, err
}

func (i *InteractiveServiceAdapter) Like(ctx context.Context, in *intrv1.LikeRequest, opts ...grpc.CallOption) (*intrv1.LikeResponse, error) {
	err := i.svc.Like(ctx, in.GetBiz(), in.GetBizId(), in.GetUid())
	return &intrv1.LikeResponse{}, err
}

func (i *InteractiveServiceAdapter) CancelLike(ctx context.Context, in *intrv1.CancelLikeRequest, opts ...grpc.CallOption) (*intrv1.CancelLikeResponse, error) {
	err := i.svc.CancelLike(ctx, in.GetBiz(), in.GetBizId(), in.GetUid())
	return &intrv1.CancelLikeResponse{}, err
}

func (i *InteractiveServiceAdapter) Collect(ctx context.Context, in *intrv1.CollectRequest, opts ...grpc.CallOption) (*intrv1.CollectResponse, error) {
	err := i.svc.Collect(ctx, in.GetBiz(), in.GetBizId(), in.GetCid(), in.GetUid())
	return &intrv1.CollectResponse{}, err
}

func (i *InteractiveServiceAdapter) CancelCollect(ctx context.Context, in *intrv1.CancelCollectRequest, opts ...grpc.CallOption) (*intrv1.CancelCollectResponse, error) {
	err := i.svc.CancelCollect(ctx, in.GetBiz(), in.GetBizId(), in.GetUid())
	return &intrv1.CancelCollectResponse{}, err
}

func (i *InteractiveServiceAdapter) Get(ctx context.Context, in *intrv1.GetRequest, opts ...grpc.CallOption) (*intrv1.GetResponse, error) {
	intr, err := i.svc.Get(ctx, in.GetBiz(), in.GetBizId(), in.GetUid())
	return &intrv1.GetResponse{
		Intr: i.toDTO(intr),
	}, err
}

func (i *InteractiveServiceAdapter) GetByIds(ctx context.Context, in *intrv1.GetByIdsRequest, opts ...grpc.CallOption) (*intrv1.GetByIdsResponse, error) {
	res, err := i.svc.GetByIds(ctx, in.GetBiz(), in.GetBizIds())
	if err != nil {
		return &intrv1.GetByIdsResponse{}, err
	}
	m := make(map[int64]*intrv1.Interactive, len(res))
	for k, v := range res {
		m[k] = i.toDTO(v)
	}
	return &intrv1.GetByIdsResponse{
		Intrs: m,
	}, nil
}

func (i *InteractiveServiceAdapter) toDTO(intr domain.Interactive) *intrv1.Interactive {
	return &intrv1.Interactive{
		Biz:        intr.Biz,
		BizId:      intr.BizId,
		ReadCnt:    intr.ReadCnt,
		LikeCnt:    intr.LikeCnt,
		CollectCnt: intr.CollectCnt,
		Liked:      intr.Liked,
		Collected:  intr.Collected,
	}
}
