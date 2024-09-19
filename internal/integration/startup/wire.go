//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"webook/internal/handler"
	ijwt "webook/internal/handler/jwt"
	"webook/internal/ioc"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
)

var thirdPartySet = wire.NewSet(
	InitDB, InitRedis,
	InitLog)

var userSvcProvider = wire.NewSet(
	dao.NewUserDao,
	cache.NewUserCache,
	repository.NewCachedUserRepository,
	service.NewUserService)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdPartySet,
		userSvcProvider,

		// dao 部分
		dao.NewGormArticleDao,

		// cache 部分
		cache.NewCodeCache,

		// repo 部分
		repository.NewCodeRepository,
		repository.NewArticleRepository,

		// Service 部分
		ioc.InitSMSService,
		service.NewCodeService,
		InitWechatService,
		service.NewArticleService,

		// handler 部分
		handler.NewUserHandler,
		handler.NewOAuth2WechatHandler,
		ijwt.NewRedisJWTHandler,
		handler.NewArticleHandler,

		// 中间件，路由等？
		//gin.Default,
		ioc.InitGin,
		ioc.InitMiddlewares,
	)
	return new(gin.Engine)
}

func InitArticleHandler() *handler.ArticleHandler {
	wire.Build(thirdPartySet,
		service.NewArticleService,
		handler.NewArticleHandler,
		repository.NewArticleRepository,
		dao.NewGormArticleDao,
	)
	return &handler.ArticleHandler{}
}
