//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/code/grpc"
	"webook/code/ioc"
	"webook/code/repository"
	"webook/code/repository/cache"
	"webook/code/service"
	"webook/pkg/grpcx"
)

func InitCodeGRPCServer() *grpcx.Server {
	wire.Build(
		ioc.InitLogger,
		ioc.InitRedis,
		ioc.InitEtcd,
		ioc.InitGRPCServer,
		grpc.NewCodeGRPCServer,
		service.NewCodeService,
		repository.NewCodeRepository,
		cache.NewCodeCache,
		ioc.InitSmsGRPCClient,
	)
	return new(grpcx.Server)
}
