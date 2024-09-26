//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
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
		article2.NewGormArticleDao,

		// cache 部分
		cache.NewCodeCache,

		// repo 部分
		repository.NewCodeRepository,
		article.NewArticleRepository,

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

func InitArticleHandler(dao article2.ArticleDao) *handler.ArticleHandler {
	wire.Build(thirdPartySet,
		service.NewArticleService,
		handler.NewArticleHandler,
		article.NewArticleRepository,
		//article2.NewGormArticleDao,
	)
	return &handler.ArticleHandler{}
}
