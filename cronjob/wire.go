//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/cronjob/grpc"
	"webook/cronjob/ioc"
	"webook/cronjob/repository"
	"webook/cronjob/repository/dao"
	"webook/cronjob/service"
	"webook/pkg/grpcx"
)

func InitCronJobGRPCServer() *grpcx.Server {
	wire.Build(
		ioc.InitLogger,
		ioc.InitDB,
		ioc.InitGRPCxServer,
		grpc.NewCronJobServiceServer,
		service.NewCronJobService,
		repository.NewPreemptCronJobRepository,
		dao.NewGORMJobDAO,
	)
	return new(grpcx.Server)
}
