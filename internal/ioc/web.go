package ioc

import (
	"context"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"strings"
	"time"
	"webook/internal/handler"
	ijwt "webook/internal/handler/jwt"
	"webook/internal/handler/middleware"
	"webook/internal/handler/middleware/logger"
	logger2 "webook/internal/pkg/logger"
	"webook/internal/pkg/ratelimit"

	//"webook/internal/pkg/gin-pulgin/middlewares/ratelimit"
	limitbuilder "webook/internal/pkg/gin-pulgin/middlewares/ratelimit"
)

func InitGin(mdls []gin.HandlerFunc, hdl *handler.UserHandler, oauth2WechatHdl *handler.OAuth2WechatHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	hdl.RegisterRoutes(server)
	oauth2WechatHdl.RegisterRoutes(server)
	return server
}

func InitMiddlewares(redisClient redis.Cmdable, jwtHandler ijwt.Handler, l logger2.LoggerV1) []gin.HandlerFunc {
	limiter := ratelimit.NewRedisSlidingWindowLimiter(redisClient, time.Minute, 1000)
	bd := logger.NewBuilder(func(ctx context.Context, al *logger.AccessLog) {
		l.Info("HTTP请求", logger2.Field{Key: "al", Value: al})
	}).AllowReqBody(true).AllowRespBody()
	// 监听配置文件
	viper.OnConfigChange(func(in fsnotify.Event) {
		ok := viper.GetBool("web.logreq")
		bd.AllowReqBody(ok)
	})
	return []gin.HandlerFunc{
		corsHdl(),
		bd.Build(),
		middleware.NewLoginJWTMiddlewareBuilder(jwtHandler).
			IgnorePaths("/users/login").
			IgnorePaths("/users/signup").
			IgnorePaths("/users/login_sms/code/send").
			IgnorePaths("/users/login_sms/code/verify").
			IgnorePaths("/users/login_sms").
			IgnorePaths("/oauth2/wechat/authurl").
			IgnorePaths("/oauth2/wechat/callback").
			IgnorePaths("/users/refresh_token").
			Build(),
		//ratelimit.NewBuilder(redisClient, time.Minute, 1000).Build(),
		limitbuilder.NewBuilder(limiter).Build(),
	}
}

func corsHdl() gin.HandlerFunc {
	return cors.New(cors.Config{
		// AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: cors.DefaultConfig().AllowMethods,
		AllowHeaders: []string{"Content-Type", "Authorization"},
		// 不加这个，前端拿不到
		ExposeHeaders:    []string{"x-jwt-token", "x-refresh-token"},
		AllowCredentials: true, // 是否允许发送Cookie，默认false
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				// 开发环境
				return true
			}
			return strings.Contains(origin, "yourcompany.com") // 允许公司域名访问
		},
		MaxAge: 12 * time.Hour,
	})
}
