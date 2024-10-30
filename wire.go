//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/interactive/events"
	repository2 "webook/interactive/repository"
	cache2 "webook/interactive/repository/cache"
	dao2 "webook/interactive/repository/dao"
	service2 "webook/interactive/service"
	article3 "webook/internal/events/article"
	"webook/internal/handler"
	ijwt "webook/internal/handler/jwt"
	"webook/internal/ioc"
	"webook/internal/repository"
	"webook/internal/repository/article"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	article2 "webook/internal/repository/dao/article"
	"webook/internal/service"
)

// User 相关依赖
var UserSet = wire.NewSet(
	handler.NewUserHandler,
	service.NewUserService,
	dao.NewUserDao,
	cache.NewUserCache,
	repository.NewCachedUserRepository,
)

// Gorm 文章相关依赖
var GormArticleSet = wire.NewSet(
	handler.NewArticleHandler,
	service.NewArticleService,
	article.NewArticleRepository,
	article2.NewGormArticleDao,
	article2.InitCollections,
	cache.NewRedisArticleCache,
)

// Mongo 文章相关依赖
var MongoArticleSet = wire.NewSet(
	ioc.InitMongoDB,
	ioc.InitSnowflakeNode,
	handler.NewArticleHandler,
	service.NewArticleService,
	article.NewArticleRepository,
	article2.NewMongoArticleDao,
)

// S3 文章相关依赖：将制作库存储所有信息，线上库存储除文章以外的信息，oss存储文章
var S3ArticleSet = wire.NewSet(
	handler.NewArticleHandler,
	service.NewArticleService,
	article.NewArticleRepository,
	article2.NewOssDAO,
	ioc.InitOss,
)

// 短信相关依赖
var CodeSet = wire.NewSet(
	ioc.InitSMSService,
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
