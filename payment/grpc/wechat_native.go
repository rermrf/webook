package grpc

import (
	"context"
	"google.golang.org/grpc"
	pmtv1 "webook/api/proto/gen/payment/v1"
	"webook/payment/domain"
	"webook/payment/service/wechat"
)

type WechatServiceServer struct {
	pmtv1.UnimplementedWechatPaymentServiceServer
	svc *wechat.NativePaymentService
}

func NewWechatServiceServer(svc *wechat.NativePaymentService) *WechatServiceServer {
	return &WechatServiceServer{svc: svc}
}

func (w *WechatServiceServer) Register(server *grpc.Server) {
	pmtv1.RegisterWechatPaymentServiceServer(server, w)
}

func (w *WechatServiceServer) NativePrePay(ctx context.Context, request *pmtv1.PrePayRequest) (*pmtv1.NativePrePayResponse, error) {
	codeUrl, err := w.svc.Prepay(ctx, domain.Payment{
		Amt: domain.Amount{
			Currency: request.Amt.Currency,
			Total:    request.Amt.Total,
		},
		BizTradeNO:  request.BizTradeNo,
		Description: request.Description,
	})
	if err != nil {
		return nil, err
	}
	return &pmtv1.NativePrePayResponse{
		CodeUrl: codeUrl,
	}, nil
}

func (w *WechatServiceServer) GetPayment(ctx context.Context, request *pmtv1.GetPaymentRequest) (*pmtv1.GetPaymentResponse, error) {
	p, err := w.svc.GetPayment(ctx, request.GetBizTradeNo())
	if err != nil {
		return nil, err
	}
	return &pmtv1.GetPaymentResponse{
		Status: pmtv1.PaymentStatus(p.Status),
	}, nil
}
