package grpc

import (
	"context"
	"google.golang.org/grpc"
	rewardv1 "webook/api/proto/gen/reward/v1"
	"webook/reward/domain"
	"webook/reward/service"
)

type RewardServiceServer struct {
	svc service.RewardService
	rewardv1.UnimplementedRewardServiceServer
}

func NewRewardServiceServer(svc service.RewardService) *RewardServiceServer {
	return &RewardServiceServer{svc: svc}
}

func (r *RewardServiceServer) Register(server *grpc.Server) {
	rewardv1.RegisterRewardServiceServer(server, r)
}

func (r *RewardServiceServer) PreReward(ctx context.Context, request *rewardv1.PreRewardRequest) (*rewardv1.PreRewardResponse, error) {
	codeURL, err := r.svc.PreReward(ctx, domain.Reward{
		Uid: request.GetUid(),
		Target: domain.Target{
			Biz:     request.Biz,
			BizId:   request.BizId,
			BizName: request.BizName,
			Uid:     request.GetUid(),
		},
		Amt: request.GetAmt(),
	})
	return &rewardv1.PreRewardResponse{
		CodeUrl: codeURL.URL,
		Rid:     codeURL.Rid,
	}, err
}

func (r *RewardServiceServer) GetReward(ctx context.Context, request *rewardv1.GetRewardRequest) (*rewardv1.GetRewardResponse, error) {
	rw, err := r.svc.GetReward(ctx, request.GetRid(), request.GetUid())
	if err != nil {
		return nil, err
	}
	return &rewardv1.GetRewardResponse{
		// 两个的取值是一样的，所以可以直接转
		Status: rewardv1.RewardStatus(rw.Status),
	}, nil
}
