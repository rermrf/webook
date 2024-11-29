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
	"webook/bff/handler"
	ijwt "webook/bff/handler/jwt"
	"webook/bff/ioc"
	repository2 "webook/interactive/repository"
	cache2 "webook/interactive/repository/cache"
	dao2 "webook/interactive/repository/dao"
	service2 "webook/interactive/service"
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

var clientSet = wire.NewSet(
	InitArticleGRPCClient,
	InitIntrGRPCClient,
	InitOAuth2GRPCClient,
	InitCodeGRPCClient,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdPartySet,
		userSvcProvider,
		articleSet,
		interactiveSet,

		// handler 部分
		handler.NewOAuth2WechatHandler,
		ijwt.NewRedisJWTHandler,
		clientSet,

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
		handler.NewArticleHandler,
		InitIntrGRPCClient,
		InitArticleGRPCClient,
	)
	return &handler.ArticleHandler{}
}
