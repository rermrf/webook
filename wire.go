//go:build wireinject

package main

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

var UserSet = wire.NewSet(
	handler.NewUserHandler,
	dao.NewUserDao,
	cache.NewUserCache,
	repository.NewCachedUserRepository,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		ioc.InitDB, ioc.InitRedis, ioc.InitLogger,
		cache.NewCodeCache, article2.NewGormArticleDao,
		repository.NewCodeRepository, article.NewArticleRepository,
		service.NewUserService, service.NewCodeService, service.NewArticleService, ioc.InitSMSService,
		handler.NewOAuth2WechatHandler,
		handler.NewArticleHandler,
		ijwt.NewRedisJWTHandler,
		// 中间件，路由等？
		//gin.Default,
		ioc.InitGin,
		ioc.InitMiddlewares,
		ioc.InitOAuth2WechatService,
		//UserSet,
		handler.NewUserHandler,
		dao.NewUserDao,
		cache.NewUserCache,
		repository.NewCachedUserRepository,
	)
	return new(gin.Engine)
}
