//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"webook/article/events"
	repository4 "webook/article/repository"
	cache4 "webook/article/repository/cache"
	article2 "webook/article/repository/dao"
	service3 "webook/article/service"
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
	repository3 "webook/user/repository"
	cache3 "webook/user/repository/cache"
	"webook/user/repository/dao"
)

var thirdPartySet = wire.NewSet(
	NewSyncProducer,
	InitKafka,
	InitDB,
	InitRedis,
	InitLog,
)

var userSvcProvider = wire.NewSet(
	dao.NewUserDao,
	cache3.NewUserCache,
	repository3.NewCachedUserRepository,
	InitUserGRPCClient,
	handler.NewUserHandler,
)

var articleSet = wire.NewSet(
	handler.NewArticleHandler,
	service3.NewArticleService,
	repository4.NewArticleRepository,
	article2.NewGormArticleDao,
	cache4.NewRedisArticleCache,
	events.NewKafkaProducer,
)

var interactiveSet = wire.NewSet(
	service2.NewInteractiveService,
	repository2.NewCachedInteractiveRepository,
	dao2.NewGORMInteractiveDao,
	cache2.NewRedisInteractiveCache,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdPartySet,
		userSvcProvider,
		articleSet,
		interactiveSet,

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
		InitIntrGRPCClient,

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
		interactiveSet,
		service3.NewArticleService,
		handler.NewArticleHandler,
		repository4.NewArticleRepository,
		//article2.NewGormArticleDao,
		cache4.NewRedisArticleCache,
		//dao.NewUserDao,
		//cache3.NewUserCache,
		//repository3.NewCachedUserRepository,
		events.NewKafkaProducer,
		InitIntrGRPCClient,
	)
	return &handler.ArticleHandler{}
}
