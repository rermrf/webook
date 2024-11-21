//go:build wireinject

package main

import (
	"github.com/google/wire"
	igrpc "webook/account/grpc"
	"webook/account/ioc"
	"webook/account/repository"
	"webook/account/repository/dao"
	"webook/account/service"
	"webook/pkg/grpcx"
)

func InitGRPCServiceServer() *grpcx.Server {
	wire.Build(
		ioc.InitDB,
		//ioc.InitEtcd,
		ioc.InitLogger,
		ioc.InitGRPCServer,
		igrpc.NewAccountServiceServer,
		service.NewAccountService,
		repository.NewAccountRepository,
		dao.NewAccountGORMDAO,
	)
	return &grpcx.Server{}
}
