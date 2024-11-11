//go:build wireinject

package main

import (
	"github.com/google/wire"
	article3 "webook/article/events"
	"webook/interactive/events"
	repository2 "webook/interactive/repository"
	cache2 "webook/interactive/repository/cache"
	dao2 "webook/interactive/repository/dao"
	service2 "webook/interactive/service"
	"webook/internal/handler"
	ijwt "webook/internal/handler/jwt"
	"webook/internal/ioc"
)

// User 相关依赖
var UserSet = wire.NewSet(
	handler.NewUserHandler,
)

// Gorm 文章相关依赖
var GormArticleSet = wire.NewSet(
	handler.NewArticleHandler,
)

var ThirdPartySet = wire.NewSet(
	ioc.InitRedis,
	ioc.InitDB,
	ioc.InitLogger,
	ijwt.NewRedisJWTHandler,
	ioc.InitEtcd,
)

var InteractiveSet = wire.NewSet(
	service2.NewInteractiveService,
	repository2.NewCachedInteractiveRepository,
	dao2.NewGORMInteractiveDao,
	cache2.NewRedisInteractiveCache,
	events.NewInteractiveReadBatchConsumer,
)

var OAuth2Set = wire.NewSet(
	handler.NewOAuth2WechatHandler,
)

var KafkaSet = wire.NewSet(
	ioc.InitKafka,
	ioc.NewConsumer,
	ioc.NewSyncProducer,
	article3.NewKafkaProducer,
)

var grpcClientSet = wire.NewSet(
	ioc.InitIntrGRPCClientV2,
	ioc.InitUserGRPCClient,
	ioc.InitArticleGRPCClientV1,
	ioc.InitSMSGRPCClient,
	ioc.InitCodeGRPCClient,
	ioc.InitRankingGRPCClient,
	ioc.InitOAuth2GRPCClient,
)

func InitApp() *App {
	wire.Build(
		// 中间件，路由等？
		//gin.Default,
		ioc.InitGin,
		ioc.InitMiddlewares,
		UserSet,

		//CodeSet,
		ThirdPartySet,
		OAuth2Set,
		InteractiveSet,
		grpcClientSet,
		KafkaSet,
		GormArticleSet,
		// 组装我这个结构体的所有字段

		ioc.InitRankingJob,
		ioc.InitJob,
		ioc.InitRLockClient,

		wire.Struct(new(App), "*"),
	)
	return new(App)
}
