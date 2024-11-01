//go:build wireinject

package main

import (
	"github.com/google/wire"
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
)

func InitUserGRPCServer() *grpc.UserGRPCServer {
	wire.Build(
		thirdPartySet,
		userSet,
	)
	return new(grpc.UserGRPCServer)
}
