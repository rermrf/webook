package ioc

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
	"webook/internal/handler"
	"webook/internal/handler/middleware"
	"webook/internal/pkg/gin-pulgin/middlewares/ratelimit"
)

func InitGin(mdls []gin.HandlerFunc, hdl *handler.UserHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	hdl.RegisterRoutes(server)
	return server
}

func InitMiddlewares(redisClient redis.Cmdable) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		corsHdl(),
		middleware.NewLoginJWTMiddlewareBuilder().
			IgnorePaths("/users/login").
			IgnorePaths("/users/signup").
			IgnorePaths("/users/login_sms/code/send").
			IgnorePaths("/users/login_sms/code/verify").
			IgnorePaths("/users/login_sms").
			Build(),
		ratelimit.NewBuilder(redisClient, time.Minute, 1000).Build(),
	}
}

func corsHdl() gin.HandlerFunc {
	return cors.New(cors.Config{
		// AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: cors.DefaultConfig().AllowMethods,
		AllowHeaders: []string{"Content-Type", "Authorization"},
		// 不加这个，前端拿不到
		ExposeHeaders:    []string{"x-jwt-token"},
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
