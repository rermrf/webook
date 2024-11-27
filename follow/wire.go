//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/follow/grpc"
	"webook/follow/ioc"
	"webook/follow/repository"
	"webook/follow/repository/cache"
	"webook/follow/repository/dao"
	"webook/follow/service"
	"webook/pkg/grpcx"
)

var thirdPartySet = wire.NewSet(
	ioc.InitDB,
	ioc.InitRedis,
	ioc.InitLogger,
	ioc.InitEtcd,
)

func InitFollowGRPCServer() *grpcx.Server {
	wire.Build(
		thirdPartySet,
		ioc.InitGRPCServer,
		grpc.NewFollowServiceServer,
		service.NewFollowService,
		repository.NewCachedFollowRepository,
		cache.NewRedisFollowCache,
		dao.NewGORMFollowDao,
	)
	return new(grpcx.Server)
}
