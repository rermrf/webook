//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/feed/events"
	"webook/feed/grpc"
	"webook/feed/ioc"
	"webook/feed/repository"
	"webook/feed/repository/cache"
	"webook/feed/repository/dao"
	"webook/feed/service"
	"webook/pkg/app"
)

var thirdPartyWireSet = wire.NewSet(
	ioc.InitEtcd,
	ioc.InitRedis,
	ioc.InitDB,
	ioc.InitLogger,
	ioc.InitFollowClient,
	ioc.InitKafka,
)

var serviceSet = wire.NewSet(
	dao.NewFeedPullEventDao,
	dao.NewFeedPushEventDao,
	cache.NewFeedEventCache,
	repository.NewFeedEventRepository,
)

func Init() *app.App {
	wire.Build(
		thirdPartyWireSet,
		serviceSet,
		ioc.RegisterHandler,
		service.NewFeedService,
		grpc.NewFeedEventGrpcServer,
		events.NewArticleEventConsumer,
		events.NewFeedEventConsumer,
		ioc.InitGRPCServer,
		ioc.NewConsumers,
		wire.Struct(new(app.App), "*"),
	)
	return &app.App{}
}
