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
	ioc.InitDST,
	ioc.InitSRC,
	ioc.InitDoubleWritePool,
	ioc.InitBizDB,
	//ioc.InitDB,
	ioc.InitLogger,
	ioc.InitKafka,
	ioc.InitSyncProducer,
	ioc.InitRedis,
)

var interactiveSet = wire.NewSet(
	grpc.NewInteractiveServiceServer,
	service.NewInteractiveService,
	repository.NewCachedInteractiveRepository,
	dao.NewGORMInteractiveDao,
	cache.NewRedisInteractiveCache,
)

var migratorSet = wire.NewSet(
	ioc.InitMigratorServer,
	ioc.InitMigradatorProducer,
	ioc.InitFixDataConsumer,
)

func InitApp() *App {
	wire.Build(
		interactiveSet,
		thirdPartySet,
		migratorSet,
		ioc.NewConsumers,
		ioc.InitGRPCServer,
		events.NewInteractiveReadEventConsumer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
