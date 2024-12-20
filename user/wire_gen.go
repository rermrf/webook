// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

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

// Injectors from wire.go:

func InitUserGRPCServer() *grpcx.Server {
	loggerV1 := ioc.InitLogger()
	db := ioc.InitDB(loggerV1)
	userDao := dao.NewUserDao(db)
	cmdable := ioc.InitRedis()
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewCachedUserRepository(userDao, userCache)
	client := ioc.InitKafka()
	producer := ioc.InitProducer(client)
	userService := service.NewUserService(userRepository, loggerV1, producer)
	userGRPCServer := grpc.NewUserGRPCServer(userService)
	server := ioc.InitGRPCServer(userGRPCServer, loggerV1)
	return server
}

// wire.go:

var userSet = wire.NewSet(grpc.NewUserGRPCServer, service.NewUserService, repository.NewCachedUserRepository, dao.NewUserDao, cache.NewUserCache)

var thirdPartySet = wire.NewSet(ioc.InitDB, ioc.InitLogger, ioc.InitRedis, ioc.InitKafka, ioc.InitProducer)
