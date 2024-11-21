package grpc

import (
	"context"
	"google.golang.org/grpc"
	"webook/account/domain"
	"webook/account/service"
	accountv1 "webook/api/proto/gen/account/v1"
)

type AccountServiceServer struct {
	accountv1.UnimplementedAccountServiceServer
	svc service.AccountService
}

func NewAccountServiceServer(svc service.AccountService) *AccountServiceServer {
	return &AccountServiceServer{svc: svc}
}

func (a *AccountServiceServer) Register(server *grpc.Server) {
	accountv1.RegisterAccountServiceServer(server, a)
}

func (a *AccountServiceServer) Credit(ctx context.Context, request *accountv1.CreditRequest) (*accountv1.CreditResponse, error) {
	err := a.svc.Credit(ctx, a.toDomain(request))
	return &accountv1.CreditResponse{}, err
}

func (a *AccountServiceServer) toDomain(c *accountv1.CreditRequest) domain.Credit {
	return domain.Credit{
		Biz:   c.GetBiz(),
		BizId: c.GetBizId(),
		Items: a.itemToDomains(c.GetItems()),
	}
}

func (a *AccountServiceServer) itemToDomains(cis []*accountv1.CreditItem) []domain.CreditItem {
	res := make([]domain.CreditItem, len(cis))
	for i, ci := range cis {
		res[i] = domain.CreditItem{
			Uid:         ci.GetUid(),
			Account:     ci.GetAccount(),
			AccountType: domain.AccountType(ci.GetAccountType()),
			Amt:         ci.GetAmt(),
			Currency:    ci.GetCurrency(),
		}
	}
	return res
}
