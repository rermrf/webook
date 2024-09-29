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

var OAuth2Set = wire.NewSet(
	handler.NewOAuth2WechatHandler,
	ioc.InitOAuth2WechatService,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 中间件，路由等？
		//gin.Default,
		ioc.InitGin,
		ioc.InitMiddlewares,
		UserSet,
		GormArticleSet,
		//MongoArticleSet,
		//S3ArticleSet,
		CodeSet,
		ThirdPartySet,
		OAuth2Set,
	)
	return new(gin.Engine)
}
