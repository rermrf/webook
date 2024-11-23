//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/payment/grpc"
	"webook/payment/ioc"
	"webook/payment/repository"
	"webook/payment/repository/dao"
	"webook/payment/web"
)

var thirdPartySet = wire.NewSet(
	ioc.InitDB,
	ioc.InitLogger,
	ioc.InitEtcd,
	ioc.InitKafka,
)

func InitApp() App {
	wire.Build(
		thirdPartySet,
		ioc.InitGRPCServer,
		web.NewWechatHandler,
		ioc.InitWechatNativeService,
		ioc.InitProducer,
		repository.NewPaymentRepository,
		dao.NewPaymentGORMDAO,
		grpc.NewWechatServiceServer,
		ioc.InitGinServer,
		ioc.InitWechatConfig,
		ioc.InitWechatClient,
		ioc.InitWechatNotifyHandler,
		wire.Struct(new(App), "WebServer", "GRPCServer"),
	)
	return App{}
}
