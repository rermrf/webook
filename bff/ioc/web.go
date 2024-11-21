package ioc

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"strings"
	"time"
	"webook/bff/handler"
	ijwt "webook/bff/handler/jwt"
	"webook/bff/handler/middleware"
	"webook/pkg/ginx"
	"webook/pkg/ginx/middlewares/metric"
	limitbuilder "webook/pkg/ginx/middlewares/ratelimit"
	logger2 "webook/pkg/logger"
	"webook/pkg/ratelimit"
)

func InitGin(mdls []gin.HandlerFunc, hdl *handler.UserHandler, oauth2WechatHdl *handler.OAuth2WechatHandler, articleHdl *handler.ArticleHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	hdl.RegisterRoutes(server)
	oauth2WechatHdl.RegisterRoutes(server)
	articleHdl.RegisterRoutes(server)
	(&handler.ObservabilityHandler{}).RegisterRoutes(server)
	return server
}

func InitMiddlewares(redisClient redis.Cmdable, jwtHandler ijwt.Handler, l logger2.LoggerV1) []gin.HandlerFunc {
	limiter := ratelimit.NewRedisSlidingWindowLimiter(redisClient, time.Minute, 1000)
	//bd := logger.NewBuilder(func(ctx context.Context, al *logger.AccessLog) {
	//	l.Info("HTTP请求", logger2.Field{Key: "al", Value: al})
	//}).AllowReqBody(true).AllowRespBody()
	//// 监听配置文件
	//viper.OnConfigChange(func(in fsnotify.Event) {
	//	ok := viper.GetBool("web.logreq")
	//	bd.AllowReqBody(ok)
	//})

	ginx.InitCounter(prometheus.CounterOpts{
		Namespace: "emoji",
		Subsystem: "webook",
		Name:      "http_biz_code",
		Help:      "HTTP 的业务错误码",
	})
	return []gin.HandlerFunc{
		corsHdl(),
		(&metric.MiddlewareBuilder{
			Namespace:  "emoji",
			Subsystem:  "webook",
			Name:       "gin_http",
			Help:       "统计 GIN 的 HTTP 接口",
			InstanceID: "my_instance_id",
		}).Build(),
		//bd.Build(),
		middleware.NewLoginJWTMiddlewareBuilder(jwtHandler).
			IgnorePaths("/users/login").
			IgnorePaths("/users/signup").
			IgnorePaths("/users/login_sms/code/send").
			IgnorePaths("/users/login_sms/code/verify").
			IgnorePaths("/users/login_sms").
			IgnorePaths("/oauth2/wechat/authurl").
			IgnorePaths("/oauth2/wechat/callback").
			IgnorePaths("/users/refresh_token").
			IgnorePaths("/test/metrics").
			Build(),
		//ratelimit.NewBuilder(redisClient, time.Minute, 1000).Build(),
		limitbuilder.NewBuilder(limiter).Build(),
		// gin 接入 opentelemetry
		otelgin.Middleware("webook"),
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
