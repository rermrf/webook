//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/oauth2/grpc"
	"webook/oauth2/ioc"
	"webook/pkg/grpcx"
)

func InitOauth2GRPCServer() *grpcx.Server {
	wire.Build(
		ioc.InitLogger,
		ioc.InitService,
		ioc.InitGRPCxServer,
		grpc.NewOauth2ServiceServer,
	)
	return new(grpcx.Server)
}
