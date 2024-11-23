// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/google/wire"
	events2 "webook/article/events"
	"webook/bff/handler"
	"webook/bff/handler/jwt"
	"webook/bff/ioc"
	"webook/interactive/events"
	"webook/interactive/repository"
	"webook/interactive/repository/cache"
	"webook/interactive/repository/dao"
	"webook/interactive/service"
)

import (
	_ "github.com/spf13/viper/remote"
)

// Injectors from wire.go:

func InitApp() *App {
	cmdable := ioc.InitRedis()
	jwtHandler := jwt.NewRedisJWTHandler(cmdable)
	loggerV1 := ioc.InitLogger()
	v := ioc.InitMiddlewares(cmdable, jwtHandler, loggerV1)
	client := ioc.InitEtcd()
	userServiceClient := ioc.InitUserGRPCClient(client)
	codeServiceClient := ioc.InitCodeGRPCClient(client)
	userHandler := handler.NewUserHandler(userServiceClient, codeServiceClient, cmdable, jwtHandler, loggerV1)
	oauth2ServiceClient := ioc.InitOAuth2GRPCClient(client)
	oAuth2WechatHandler := handler.NewOAuth2WechatHandler(oauth2ServiceClient, userServiceClient, jwtHandler)
	articleServiceClient := ioc.InitArticleGRPCClientV1(client)
	interactiveServiceClient := ioc.InitIntrGRPCClientV2(client)
	rewardServiceClient := ioc.InitRewardGRPCClient(client)
	articleHandler := handler.NewArticleHandler(articleServiceClient, loggerV1, interactiveServiceClient, rewardServiceClient)
	engine := ioc.InitGin(v, userHandler, oAuth2WechatHandler, articleHandler)
	saramaClient := ioc.InitKafka()
	db := ioc.InitDB(loggerV1)
	interactiveDao := dao.NewGORMInteractiveDao(db)
	interactiveCache := cache.NewRedisInteractiveCache(cmdable)
	interactiveRepository := repository.NewCachedInteractiveRepository(interactiveDao, interactiveCache, loggerV1)
	interactiveReadBatchConsumer := events.NewInteractiveReadBatchConsumer(saramaClient, loggerV1, interactiveRepository)
	v2 := ioc.NewConsumer(interactiveReadBatchConsumer)
	rankingServiceClient := ioc.InitRankingGRPCClient(client)
	rlockClient := ioc.InitRLockClient(cmdable)
	rankingJob := ioc.InitRankingJob(rankingServiceClient, rlockClient, loggerV1)
	cron := ioc.InitJob(loggerV1, rankingJob)
	app := &App{
		Server:    engine,
		Consumers: v2,
		cron:      cron,
	}
	return app
}

// wire.go:

// User 相关依赖
var UserSet = wire.NewSet(handler.NewUserHandler)

// Gorm 文章相关依赖
var GormArticleSet = wire.NewSet(handler.NewArticleHandler)

var ThirdPartySet = wire.NewSet(ioc.InitRedis, ioc.InitDB, ioc.InitLogger, jwt.NewRedisJWTHandler, ioc.InitEtcd)

var InteractiveSet = wire.NewSet(service.NewInteractiveService, repository.NewCachedInteractiveRepository, dao.NewGORMInteractiveDao, cache.NewRedisInteractiveCache, events.NewInteractiveReadBatchConsumer)

var OAuth2Set = wire.NewSet(handler.NewOAuth2WechatHandler)

var KafkaSet = wire.NewSet(ioc.InitKafka, ioc.NewConsumer, ioc.NewSyncProducer, events2.NewKafkaProducer)

var grpcClientSet = wire.NewSet(ioc.InitIntrGRPCClientV2, ioc.InitUserGRPCClient, ioc.InitArticleGRPCClientV1, ioc.InitSMSGRPCClient, ioc.InitCodeGRPCClient, ioc.InitRankingGRPCClient, ioc.InitOAuth2GRPCClient, ioc.InitRewardGRPCClient)
