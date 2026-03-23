package grpc

import (
	"context"

	"google.golang.org/grpc"
	creditv1 "webook/api/proto/gen/credit/v1"
	"webook/credit/service"
)

type CreditServiceServer struct {
	svc service.CreditService
	creditv1.UnimplementedCreditServiceServer
}

func NewCreditServiceServer(svc service.CreditService) *CreditServiceServer {
	return &CreditServiceServer{svc: svc}
}

func (s *CreditServiceServer) Register(server *grpc.Server) {
	creditv1.RegisterCreditServiceServer(server, s)
}

// GetBalance 获取积分余额
func (s *CreditServiceServer) GetBalance(ctx context.Context, req *creditv1.GetBalanceRequest) (*creditv1.GetBalanceResponse, error) {
	balance, err := s.svc.GetBalance(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &creditv1.GetBalanceResponse{Balance: balance}, nil
}

// GetFlows 获取积分流水
func (s *CreditServiceServer) GetFlows(ctx context.Context, req *creditv1.GetFlowsRequest) (*creditv1.GetFlowsResponse, error) {
	flows, err := s.svc.GetFlows(ctx, req.GetUid(), int(req.GetOffset()), int(req.GetLimit()))
	if err != nil {
		return nil, err
	}

	protoFlows := make([]*creditv1.CreditFlow, 0, len(flows))
	for _, f := range flows {
		protoFlows = append(protoFlows, &creditv1.CreditFlow{
			Id:          f.Id,
			Biz:         f.Biz,
			BizId:       f.BizId,
			ChangeAmt:   f.ChangeAmt,
			Balance:     f.Balance,
			Description: f.Description,
			Ctime:       f.Ctime,
		})
	}
	return &creditv1.GetFlowsResponse{Flows: protoFlows}, nil
}

// EarnCredit 积分获取
func (s *CreditServiceServer) EarnCredit(ctx context.Context, req *creditv1.EarnCreditRequest) (*creditv1.EarnCreditResponse, error) {
	earned, balance, msg, err := s.svc.EarnCredit(ctx, req.GetUid(), req.GetBiz(), req.GetBizId())
	if err != nil {
		return nil, err
	}
	return &creditv1.EarnCreditResponse{
		Success:    earned > 0,
		EarnedAmt:  earned,
		NewBalance: balance,
		Message:    msg,
	}, nil
}

// RewardCredit 积分打赏
func (s *CreditServiceServer) RewardCredit(ctx context.Context, req *creditv1.RewardCreditRequest) (*creditv1.RewardCreditResponse, error) {
	rewardId, err := s.svc.RewardCredit(ctx, req.GetUid(), req.GetTargetUid(), req.GetBiz(), req.GetBizId(), req.GetAmt())
	if err != nil {
		return &creditv1.RewardCreditResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return &creditv1.RewardCreditResponse{
		RewardId: rewardId,
		Success:  true,
	}, nil
}

// GetCreditReward 获取打赏详情
func (s *CreditServiceServer) GetCreditReward(ctx context.Context, req *creditv1.GetCreditRewardRequest) (*creditv1.GetCreditRewardResponse, error) {
	reward, err := s.svc.GetCreditReward(ctx, req.GetRewardId(), req.GetUid())
	if err != nil {
		return nil, err
	}
	return &creditv1.GetCreditRewardResponse{
		Status: creditv1.CreditRewardStatus(reward.Status),
		Amt:    reward.Amount,
	}, nil
}

// GetDailyStatus 获取每日积分状态
func (s *CreditServiceServer) GetDailyStatus(ctx context.Context, req *creditv1.GetDailyStatusRequest) (*creditv1.GetDailyStatusResponse, error) {
	statuses, err := s.svc.GetDailyStatus(ctx, req.GetUid(), req.GetBiz())
	if err != nil {
		return nil, err
	}

	protoStatuses := make([]*creditv1.DailyStatus, 0, len(statuses))
	for _, st := range statuses {
		protoStatuses = append(protoStatuses, &creditv1.DailyStatus{
			Biz:         st.Biz,
			EarnedCount: st.EarnedCount,
			EarnedAmt:   st.EarnedAmt,
			DailyLimit:  st.DailyLimit,
			Remaining:   st.Remaining,
		})
	}
	return &creditv1.GetDailyStatusResponse{Statuses: protoStatuses}, nil
}
