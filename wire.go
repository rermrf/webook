//go:build wireinject

package main

import (
	"github.com/google/wire"
	article3 "webook/article/events"
	repository4 "webook/article/repository"
	cache4 "webook/article/repository/cache"
	dao3 "webook/article/repository/dao"
	service3 "webook/article/service"
	"webook/interactive/events"
	repository2 "webook/interactive/repository"
	cache2 "webook/interactive/repository/cache"
	dao2 "webook/interactive/repository/dao"
	service2 "webook/interactive/service"
	"webook/internal/handler"
	ijwt "webook/internal/handler/jwt"
	"webook/internal/ioc"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/service"
)

// User 相关依赖
var UserSet = wire.NewSet(
	handler.NewUserHandler,
)

// Gorm 文章相关依赖
var GormArticleSet = wire.NewSet(
	handler.NewArticleHandler,
	service3.NewArticleService,
	repository4.NewArticleRepository,
	dao3.NewGormArticleDao,
	dao3.InitCollections,
	cache4.NewRedisArticleCache,
)

// 短信相关依赖
var CodeSet = wire.NewSet(
	service.NewCodeService,
	cache.NewCodeCache,
	repository.NewCodeRepository,
)

var ThirdPartySet = wire.NewSet(
	ioc.InitRedis,
	ioc.InitDB,
	ioc.InitLogger,
	ijwt.NewRedisJWTHandler,
)

var InteractiveSet = wire.NewSet(
	service2.NewInteractiveService,
	repository2.NewCachedInteractiveRepository,
	dao2.NewGORMInteractiveDao,
	cache2.NewRedisInteractiveCache,
	//article3.NewInteractiveReadEventConsumer,
	events.NewInteractiveReadBatchConsumer,
)

var OAuth2Set = wire.NewSet(
	handler.NewOAuth2WechatHandler,
	ioc.InitOAuth2WechatService,
)

var KafkaSet = wire.NewSet(
	ioc.InitKafka,
	ioc.NewConsumer,
	ioc.NewSyncProducer,
	article3.NewKafkaProducer,
)

var rankingServiceSet = wire.NewSet(
	service.NewBatchRankingService,
	repository.NewCachedRankingRepository,
	cache.NewRankingRedisCache,
	cache.NewRankingLocalCache,
)

var grpcClientSet = wire.NewSet(
	ioc.InitIntrGRPCClient,
	ioc.InitUserGRPCClient,
	ioc.InitArticleGRPCClient,
	ioc.InitSMSGRPCClient,
)

func InitWebServer() *App {
	wire.Build(
		// 中间件，路由等？
		//gin.Default,
		ioc.InitGin,
		ioc.InitMiddlewares,
		UserSet,

		CodeSet,
		ThirdPartySet,
		OAuth2Set,
		InteractiveSet,
		grpcClientSet,
		KafkaSet,
		GormArticleSet,
		//MongoArticleSet,
		//S3ArticleSet,
		// 组装我这个结构体的所有字段

		rankingServiceSet,
		ioc.InitRankingJob,
		ioc.InitJob,
		ioc.InitRLockClient,

		wire.Struct(new(App), "*"),
	)
	return new(App)
}
