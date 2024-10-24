// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/google/wire"
	article3 "webook/internal/events/article"
	"webook/internal/handler"
	"webook/internal/handler/jwt"
	"webook/internal/ioc"
	"webook/internal/repository"
	article2 "webook/internal/repository/article"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/repository/dao/article"
	"webook/internal/service"
)

import (
	_ "github.com/spf13/viper/remote"
)

// Injectors from wire.go:

func InitWebServer() *App {
	cmdable := ioc.InitRedis()
	jwtHandler := jwt.NewRedisJWTHandler(cmdable)
	loggerV1 := ioc.InitLogger()
	v := ioc.InitMiddlewares(cmdable, jwtHandler, loggerV1)
	db := ioc.InitDB(loggerV1)
	userDao := dao.NewUserDao(db)
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewCachedUserRepository(userDao, userCache)
	userService := service.NewUserService(userRepository, loggerV1)
	codeCache := cache.NewCodeCache(cmdable)
	codeRepository := repository.NewCodeRepository(codeCache)
	smsService := ioc.InitSMSService()
	codeService := service.NewCodeService(codeRepository, smsService)
	userHandler := handler.NewUserHandler(userService, codeService, cmdable, jwtHandler, loggerV1)
	wechatService := ioc.InitOAuth2WechatService(loggerV1)
	oAuth2WechatHandler := handler.NewOAuth2WechatHandler(wechatService, userService, jwtHandler)
	articleDao := article.NewGormArticleDao(db)
	articleCache := cache.NewRedisArticleCache(cmdable)
	articleRepository := article2.NewArticleRepository(articleDao, articleCache, loggerV1, userRepository)
	client := ioc.InitKafka()
	syncProducer := ioc.NewSyncProducer(client)
	producer := article3.NewKafkaProducer(syncProducer)
	articleService := service.NewArticleService(articleRepository, loggerV1, producer)
	interactiveDao := dao.NewGORMInteractiveDao(db)
	interactiveCache := cache.NewRedisInteractiveCache(cmdable)
	interactiveRepository := repository.NewCachedInteractiveRepository(interactiveDao, interactiveCache, loggerV1)
	interactiveService := service.NewInteractiveService(interactiveRepository)
	articleHandler := handler.NewArticleHandler(articleService, loggerV1, interactiveService)
	engine := ioc.InitGin(v, userHandler, oAuth2WechatHandler, articleHandler)
	interactiveReadBatchConsumer := article3.NewInteractiveReadBatchConsumer(client, loggerV1, interactiveRepository)
	v2 := ioc.NewConsumer(interactiveReadBatchConsumer)
	rankingCache := cache.NewRankingRedisCache(cmdable)
	rankingRepository := repository.NewCachedRankingRepository(rankingCache)
	rankingService := service.NewBatchRankingService(articleService, interactiveService, rankingRepository)
	rankingJob := ioc.InitRankingJob(rankingService)
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
var UserSet = wire.NewSet(handler.NewUserHandler, service.NewUserService, dao.NewUserDao, cache.NewUserCache, repository.NewCachedUserRepository)

// Gorm 文章相关依赖
var GormArticleSet = wire.NewSet(handler.NewArticleHandler, service.NewArticleService, article2.NewArticleRepository, article.NewGormArticleDao, article.InitCollections, cache.NewRedisArticleCache)

// Mongo 文章相关依赖
var MongoArticleSet = wire.NewSet(ioc.InitMongoDB, ioc.InitSnowflakeNode, handler.NewArticleHandler, service.NewArticleService, article2.NewArticleRepository, article.NewMongoArticleDao)

// S3 文章相关依赖：将制作库存储所有信息，线上库存储除文章以外的信息，oss存储文章
var S3ArticleSet = wire.NewSet(handler.NewArticleHandler, service.NewArticleService, article2.NewArticleRepository, article.NewOssDAO, ioc.InitOss)

// 短信相关依赖
var CodeSet = wire.NewSet(ioc.InitSMSService, service.NewCodeService, cache.NewCodeCache, repository.NewCodeRepository)

var ThirdPartySet = wire.NewSet(ioc.InitRedis, ioc.InitDB, ioc.InitLogger, jwt.NewRedisJWTHandler)

var InteractiveSet = wire.NewSet(service.NewInteractiveService, repository.NewCachedInteractiveRepository, dao.NewGORMInteractiveDao, cache.NewRedisInteractiveCache, article3.NewInteractiveReadBatchConsumer)

var OAuth2Set = wire.NewSet(handler.NewOAuth2WechatHandler, ioc.InitOAuth2WechatService)

var KafkaSet = wire.NewSet(ioc.InitKafka, ioc.NewConsumer, ioc.NewSyncProducer, article3.NewKafkaProducer)

var rankingServiceSet = wire.NewSet(service.NewBatchRankingService, repository.NewCachedRankingRepository, cache.NewRankingRedisCache)
