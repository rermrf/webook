// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"webook/article/events"
	repository3 "webook/article/repository"
	cache3 "webook/article/repository/cache"
	dao2 "webook/article/repository/dao"
	service2 "webook/article/service"
	"webook/interactive/repository"
	"webook/interactive/repository/cache"
	"webook/interactive/repository/dao"
	"webook/interactive/service"
	"webook/internal/handler"
	"webook/internal/handler/jwt"
	"webook/internal/ioc"
	repository2 "webook/user/repository"
	cache2 "webook/user/repository/cache"
	dao3 "webook/user/repository/dao"
)

// Injectors from wire.go:

func InitWebServer() *gin.Engine {
	cmdable := InitRedis()
	jwtHandler := jwt.NewRedisJWTHandler(cmdable)
	loggerV1 := InitLog()
	v := ioc.InitMiddlewares(cmdable, jwtHandler, loggerV1)
	userServiceClient := InitUserGRPCClient()
	codeServiceClient := InitCodeGRPCClient()
	userHandler := handler.NewUserHandler(userServiceClient, codeServiceClient, cmdable, jwtHandler, loggerV1)
	oauth2ServiceClient := InitOAuth2GRPCClient()
	oAuth2WechatHandler := handler.NewOAuth2WechatHandler(oauth2ServiceClient, userServiceClient, jwtHandler)
	articleServiceClient := InitArticleGRPCClient()
	db := InitDB()
	interactiveDao := dao.NewGORMInteractiveDao(db)
	interactiveCache := cache.NewRedisInteractiveCache(cmdable)
	interactiveRepository := repository.NewCachedInteractiveRepository(interactiveDao, interactiveCache, loggerV1)
	interactiveService := service.NewInteractiveService(interactiveRepository)
	interactiveServiceClient := InitIntrGRPCClient(interactiveService)
	articleHandler := handler.NewArticleHandler(articleServiceClient, loggerV1, interactiveServiceClient)
	engine := ioc.InitGin(v, userHandler, oAuth2WechatHandler, articleHandler)
	return engine
}

func InitArticleHandler(d dao2.ArticleDao) *handler.ArticleHandler {
	articleServiceClient := InitArticleGRPCClient()
	loggerV1 := InitLog()
	db := InitDB()
	interactiveDao := dao.NewGORMInteractiveDao(db)
	cmdable := InitRedis()
	interactiveCache := cache.NewRedisInteractiveCache(cmdable)
	interactiveRepository := repository.NewCachedInteractiveRepository(interactiveDao, interactiveCache, loggerV1)
	interactiveService := service.NewInteractiveService(interactiveRepository)
	interactiveServiceClient := InitIntrGRPCClient(interactiveService)
	articleHandler := handler.NewArticleHandler(articleServiceClient, loggerV1, interactiveServiceClient)
	return articleHandler
}

// wire.go:

var thirdPartySet = wire.NewSet(
	NewSyncProducer,
	InitKafka,
	InitDB,
	InitRedis,
	InitLog,
)

var userSvcProvider = wire.NewSet(dao3.NewUserDao, cache2.NewUserCache, repository2.NewCachedUserRepository, InitUserGRPCClient, handler.NewUserHandler)

var articleSet = wire.NewSet(handler.NewArticleHandler, service2.NewArticleService, repository3.NewArticleRepository, dao2.NewGormArticleDao, cache3.NewRedisArticleCache, events.NewKafkaProducer)

var interactiveSet = wire.NewSet(service.NewInteractiveService, repository.NewCachedInteractiveRepository, dao.NewGORMInteractiveDao, cache.NewRedisInteractiveCache)

var clientSet = wire.NewSet(
	InitArticleGRPCClient,
	InitIntrGRPCClient,
	InitOAuth2GRPCClient,
	InitCodeGRPCClient,
)
