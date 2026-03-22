//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/history/grpc"
	"webook/history/ioc"
	"webook/history/repository"
	"webook/history/repository/dao"
	"webook/history/service"
)

var thirdPartySet = wire.NewSet(
	ioc.InitDB,
	ioc.InitLogger,
)

var historySet = wire.NewSet(
	grpc.NewHistoryServiceServer,
	service.NewHistoryService,
	repository.NewHistoryRepository,
	dao.NewGORMHistoryDAO,
)

func InitApp() *App {
	wire.Build(
		thirdPartySet,
		historySet,
		ioc.InitGRPCServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
