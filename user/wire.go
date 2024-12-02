//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/pkg/grpcx"
	"webook/user/grpc"
	"webook/user/ioc"
	"webook/user/repository"
	"webook/user/repository/cache"
	"webook/user/repository/dao"
	"webook/user/service"
)

var userSet = wire.NewSet(
	grpc.NewUserGRPCServer,
	service.NewUserService,
	repository.NewCachedUserRepository,
	dao.NewUserDao,
	cache.NewUserCache,
)

var thirdPartySet = wire.NewSet(
	ioc.InitDB,
	ioc.InitLogger,
	ioc.InitRedis,
	ioc.InitKafka,
	ioc.InitProducer,
)

func InitUserGRPCServer() *grpcx.Server {
	wire.Build(
		thirdPartySet,
		userSet,
		ioc.InitGRPCServer,
	)
	return new(grpcx.Server)
}
