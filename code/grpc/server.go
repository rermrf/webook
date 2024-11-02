package grpc

import (
	"context"
	"google.golang.org/grpc"
	codev1 "webook/api/proto/gen/code/v1"
	"webook/code/service"
)

type CodeGRPCServer struct {
	svc service.CodeService
	codev1.UnimplementedCodeServiceServer
}

func NewCodeGRPCServer(svc service.CodeService) *CodeGRPCServer {
	return &CodeGRPCServer{svc: svc}
}

func (c *CodeGRPCServer) Register(server grpc.ServiceRegistrar) {
	codev1.RegisterCodeServiceServer(server, c)
}

func (c *CodeGRPCServer) Send(ctx context.Context, request *codev1.SendRequest) (*codev1.SendResponse, error) {
	err := c.svc.Send(ctx, request.GetBiz(), request.GetPhone())
	return &codev1.SendResponse{}, err
}

func (c *CodeGRPCServer) Verify(ctx context.Context, request *codev1.VerifyRequest) (*codev1.VerifyResponse, error) {
	ok, err := c.svc.Verify(ctx, request.GetBiz(), request.GetInputCode(), request.GetPhone())
	return &codev1.VerifyResponse{Answer: ok}, err
}
