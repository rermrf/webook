package grpc

import (
	"context"
	"google.golang.org/grpc"
	smsv1 "webook/api/proto/gen/sms/v1"
	"webook/sms/service"
)

type SMSGRPCServer struct {
	svc service.Service
	smsv1.UnimplementedSMSServiceServer
}

func NewSMSGRPCServer(svc service.Service) *SMSGRPCServer {
	return &SMSGRPCServer{svc: svc}
}

func (s *SMSGRPCServer) Register(server grpc.ServiceRegistrar) {
	smsv1.RegisterSMSServiceServer(server, s)
}

func (s *SMSGRPCServer) Send(ctx context.Context, request *smsv1.SendRequest) (*smsv1.SendResponse, error) {
	err := s.svc.Send(ctx, request.GetBiz(), request.GetArgs(), request.GetNumbers()...)
	return &smsv1.SendResponse{}, err
}
