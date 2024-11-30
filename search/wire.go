//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/search/events"
	"webook/search/grpc"
	"webook/search/ioc"
	"webook/search/repository"
	"webook/search/repository/dao"
	"webook/search/service"
)

var thirdPartySet = wire.NewSet(
	ioc.InitLogger,
	ioc.InitESClient,
	ioc.InitEtcd,
	ioc.InitKafka,
)

func InitApp() *App {
	wire.Build(
		thirdPartySet,
		ioc.InitGRPCServer,
		grpc.NewSearchServiceServer,
		service.NewSearchService,
		grpc.NewSyncServiceServer,
		service.NewSyncService,
		repository.NewArticleRepository,
		repository.NewAnyRepository,
		repository.NewUserRepository,
		dao.NewAnyESDao,
		dao.NewArticleESDao,
		dao.NewUserESDao,
		dao.NewTagESDao,
		events.NewArticleConsumer,
		events.NewUserConsumer,
		ioc.NewConsumers,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
