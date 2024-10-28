//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	events "webook/internal/events/article"
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

var thirdPartySet = wire.NewSet(
	NewSyncProducer,
	InitKafka,
	InitDB, InitRedis,
	InitLog,
)

var userSvcProvider = wire.NewSet(
	dao.NewUserDao,
	cache.NewUserCache,
	repository.NewCachedUserRepository,
	service.NewUserService,
	handler.NewUserHandler)

var interactiveSet = wire.NewSet(
	service.NewInteractiveService,
	repository.NewCachedInteractiveRepository,
	dao.NewGORMInteractiveDao,
	cache.NewRedisInteractiveCache,
	events.NewKafkaProducer,
)

var articleSet = wire.NewSet(
	handler.NewArticleHandler,
	service.NewArticleService,
	article.NewArticleRepository,
	article2.NewGormArticleDao,
	cache.NewRedisArticleCache,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdPartySet,
		userSvcProvider,
		interactiveSet,
		articleSet,

		// cache 部分
		cache.NewCodeCache,

		// repo 部分
		repository.NewCodeRepository,

		// Service 部分
		ioc.InitSMSService,
		service.NewCodeService,
		InitWechatService,

		// handler 部分
		handler.NewOAuth2WechatHandler,
		ijwt.NewRedisJWTHandler,

		// 中间件，路由等？
		//gin.Default,
		ioc.InitGin,
		ioc.InitMiddlewares,
	)
	return new(gin.Engine)
}

func InitArticleHandler(d article2.ArticleDao) *handler.ArticleHandler {
	wire.Build(
		thirdPartySet,
		service.NewArticleService,
		handler.NewArticleHandler,
		article.NewArticleRepository,
		//article2.NewGormArticleDao,
		cache.NewRedisArticleCache,
		dao.NewUserDao,
		cache.NewUserCache,
		repository.NewCachedUserRepository,
		events.NewKafkaProducer,

		cache.NewRedisInteractiveCache,
		dao.NewGORMInteractiveDao,
		repository.NewCachedInteractiveRepository,
		service.NewInteractiveService,
	)
	return &handler.ArticleHandler{}
}

func InitInteractiveService() service.InteractiveService {
	wire.Build(
		thirdPartySet,
		interactiveSet,
	)
	return service.NewInteractiveService(nil)
}
