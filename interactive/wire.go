//go:build wireinject

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

var thirdPartySet = wire.NewSet(
	ioc.InitDB,
	ioc.InitLogger,
	ioc.InitKafka,
	ioc.InitRedis,
)

var interactiveSet = wire.NewSet(
	grpc.NewInteractiveServiceServer,
	service.NewInteractiveService,
	repository.NewCachedInteractiveRepository,
	dao.NewGORMInteractiveDao,
	cache.NewRedisInteractiveCache,
)

func InitApp() *App {
	wire.Build(
		interactiveSet,
		thirdPartySet,
		ioc.NewConsumers,
		ioc.InitGRPCServer,
		events.NewInteractiveReadEventConsumer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
