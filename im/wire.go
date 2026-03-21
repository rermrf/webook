//go:build wireinject

package main

import (
	"github.com/google/wire"

	grpc2 "webook/im/grpc"
	"webook/im/ioc"
	"webook/im/repository"
	"webook/im/repository/cache"
	"webook/im/repository/dao"
	"webook/im/service"
)

var thirdPartySet = wire.NewSet(
	ioc.InitMongo,
	ioc.InitLogger,
	ioc.InitRedis,
)

var daoSet = wire.NewSet(
	dao.NewMessageDAO,
	dao.NewConversationDAO,
)

var cacheSet = wire.NewSet(
	cache.NewRedisIMCache,
)

var repoSet = wire.NewSet(
	repository.NewMessageRepository,
	repository.NewConversationRepository,
)

var serviceSet = wire.NewSet(
	service.NewMessageService,
	service.NewConversationService,
)

var serverSet = wire.NewSet(
	grpc2.NewIMServiceServer,
	ioc.InitGRPCServer,
)

func InitApp() *App {
	wire.Build(
		thirdPartySet,
		daoSet,
		cacheSet,
		repoSet,
		serviceSet,
		serverSet,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
