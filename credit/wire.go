//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/credit/events"
	"webook/credit/grpc"
	"webook/credit/ioc"
	"webook/credit/repository"
	"webook/credit/repository/cache"
	"webook/credit/repository/dao"
	"webook/credit/service"
)

var thirdPartySet = wire.NewSet(
	ioc.InitDB,
	ioc.InitEtcd,
	ioc.InitLogger,
	ioc.InitRedis,
	ioc.InitKafka,
)

var creditSet = wire.NewSet(
	grpc.NewCreditServiceServer,
	service.NewCreditService,
	repository.NewCreditRepository,
	dao.NewCreditGORMDAO,
	cache.NewCreditRedisCache,
)

var openAPISet = wire.NewSet(
	ioc.InitOpenAPIGRPCClient,
	ioc.InitCreditAPIHandler,
	ioc.InitEpayHandler,
)

var consumerSet = wire.NewSet(
	events.NewReadEventConsumer,
	events.NewLikeEventConsumer,
	events.NewCollectEventConsumer,
)

func InitApp() *App {
	wire.Build(
		thirdPartySet,
		creditSet,
		openAPISet,
		consumerSet,
		ioc.NewConsumers,
		ioc.InitGRPCServer,
		ioc.InitHTTPServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
