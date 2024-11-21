//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/pkg/grpcx"
	grpc2 "webook/reward/grpc"
	"webook/reward/ioc"
	"webook/reward/repository"
	"webook/reward/repository/cache"
	"webook/reward/repository/dao"
	"webook/reward/service"
)

var thirdPartyWireSet = wire.NewSet(
	ioc.InitDB,
	ioc.InitEtcd,
	ioc.InitLogger,
	ioc.InitRedis,
)

func InitRewardGRPCServer() *grpcx.Server {
	wire.Build(
		thirdPartyWireSet,
		ioc.InitGRPCServer,
		ioc.InitPaymentGRPCClientV1,
		ioc.InitAccountGRPCClientV1,
		grpc2.NewRewardServiceServer,
		service.NewWechatNativeRewardService,
		repository.NewRewardRepository,
		dao.NewRewardGORMDAO,
		cache.NewRewardRedisCache,
	)
	return new(grpcx.Server)
}
