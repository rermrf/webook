//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/pkg/grpcx"
	"webook/sms/grpc"
	"webook/sms/ioc"
)

func InitSMSGRPCServer() *grpcx.Server {
	wire.Build(
		ioc.InitLogger,
		ioc.InitGRPCServer,
		ioc.InitSMSService,
		grpc.NewSMSGRPCServer,
	)
	return new(grpcx.Server)
}
