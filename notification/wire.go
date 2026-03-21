//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/notification/events"
	"webook/notification/grpc"
	"webook/notification/ioc"
	"webook/notification/repository"
	"webook/notification/repository/cache"
	"webook/notification/repository/dao"
	"webook/notification/service"
	"webook/notification/service/channel"
)

var thirdPartySet = wire.NewSet(
	ioc.InitDB,
	ioc.InitLogger,
	ioc.InitRedis,
	ioc.InitKafka,
	ioc.InitSyncProducer,
	ioc.InitETCDClient,
)

var templateSet = wire.NewSet(
	service.NewTemplateService,
	repository.NewCachedTemplateRepository,
	dao.NewGORMTemplateDAO,
	cache.NewRedisTemplateCache,
)

var channelSet = wire.NewSet(
	channel.NewInAppSender,
	channel.NewSMSSender,
	channel.NewEmailSender,
	ioc.InitSMSProvider,
	ioc.InitChannelSenders,
)

var notificationSet = wire.NewSet(
	grpc.NewNotificationServiceServer,
	service.NewNotificationService,
	repository.NewCachedNotificationRepository,
	repository.NewTransactionRepository,
	dao.NewGORMNotificationDAO,
	dao.NewGORMTransactionDAO,
	cache.NewRedisNotificationCache,
)

var schedulerSet = wire.NewSet(
	ioc.InitCheckBackScheduler,
	ioc.InitScheduledSendJob,
	ioc.InitCronJobs,
)

var consumerSet = wire.NewSet(
	events.NewNotificationEventConsumer,
	events.NewLikeEventConsumer,
	events.NewCollectEventConsumer,
	events.NewCommentEventConsumer,
	events.NewFollowEventConsumer,
	ioc.NewConsumers,
)

func InitApp() *App {
	wire.Build(
		thirdPartySet,
		templateSet,
		channelSet,
		notificationSet,
		schedulerSet,
		consumerSet,
		ioc.InitGRPCServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
