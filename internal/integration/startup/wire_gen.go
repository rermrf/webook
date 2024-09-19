// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"webook/internal/handler"
	"webook/internal/handler/jwt"
	"webook/internal/ioc"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
)

// Injectors from wire.go:

func InitWebServer() *gin.Engine {
	cmdable := InitRedis()
	jwtHandler := jwt.NewRedisJWTHandler(cmdable)
	loggerV1 := InitLog()
	v := ioc.InitMiddlewares(cmdable, jwtHandler, loggerV1)
	db := InitDB()
	userDao := dao.NewUserDao(db)
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewCachedUserRepository(userDao, userCache)
	userService := service.NewUserService(userRepository, loggerV1)
	codeCache := cache.NewCodeCache(cmdable)
	codeRepository := repository.NewCodeRepository(codeCache)
	smsService := ioc.InitSMSService()
	codeService := service.NewCodeService(codeRepository, smsService)
	userHandler := handler.NewUserHandler(userService, codeService, cmdable, jwtHandler)
	wechatService := InitWechatService(loggerV1)
	oAuth2WechatHandler := handler.NewOAuth2WechatHandler(wechatService, userService, jwtHandler)
	articleDao := dao.NewGormArticleDao(db)
	articleRepository := repository.NewArticleRepository(articleDao)
	articleService := service.NewArticleService(articleRepository)
	articleHandler := handler.NewArticleHandler(articleService, loggerV1)
	engine := ioc.InitGin(v, userHandler, oAuth2WechatHandler, articleHandler)
	return engine
}

func InitArticleHandler() *handler.ArticleHandler {
	db := InitDB()
	articleDao := dao.NewGormArticleDao(db)
	articleRepository := repository.NewArticleRepository(articleDao)
	articleService := service.NewArticleService(articleRepository)
	loggerV1 := InitLog()
	articleHandler := handler.NewArticleHandler(articleService, loggerV1)
	return articleHandler
}

// wire.go:

var thirdPartySet = wire.NewSet(
	InitDB, InitRedis,
	InitLog)

var userSvcProvider = wire.NewSet(dao.NewUserDao, cache.NewUserCache, repository.NewCachedUserRepository, service.NewUserService)
