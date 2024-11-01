// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/google/wire"
	"webook/interactive/events"
	"webook/interactive/grpc"
	"webook/interactive/ioc"
	"webook/interactive/repository"
	"webook/interactive/repository/cache"
	"webook/interactive/repository/dao"
	"webook/interactive/service"
)

// Injectors from wire.go:

func InitApp() *App {
	loggerV1 := ioc.InitLogger()
	db := ioc.InitDB(loggerV1)
	interactiveDao := dao.NewGORMInteractiveDao(db)
	cmdable := ioc.InitRedis()
	interactiveCache := cache.NewRedisInteractiveCache(cmdable)
	interactiveRepository := repository.NewCachedInteractiveRepository(interactiveDao, interactiveCache, loggerV1)
	interactiveService := service.NewInteractiveService(interactiveRepository)
	interactiveServiceServer := grpc.NewInteractiveServiceServer(interactiveService)
	server := ioc.InitGRPCServer(interactiveServiceServer)
	client := ioc.InitKafka()
	interactiveReadEventConsumer := events.NewInteractiveReadEventConsumer(client, interactiveRepository, loggerV1)
	v := ioc.NewConsumers(interactiveReadEventConsumer)
	app := &App{
		server:    server,
		consumers: v,
	}
	return app
}

// wire.go:

var thirdPartySet = wire.NewSet(ioc.InitDB, ioc.InitLogger, ioc.InitKafka, ioc.InitRedis)

var interactiveSet = wire.NewSet(grpc.NewInteractiveServiceServer, service.NewInteractiveService, repository.NewCachedInteractiveRepository, dao.NewGORMInteractiveDao, cache.NewRedisInteractiveCache)