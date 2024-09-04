package main

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
	"webook/config"
	handler "webook/internal/handler"
	"webook/internal/handler/middleware"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/service/sms/memory"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	db := initDB()
	rdb := initRDB()
	server := initWebServer()
	u := initUser(db, rdb)
	u.RegisterRoutes(server)

	err := server.Run(config.Config.Server.HTTPPort)
	if err != nil {
		return
	}
}

func initRDB() *redis.Client {
	return redis.NewClient(&redis.Options{Addr: config.Config.Redis.Addr})
}

func initDB() *gorm.DB {
	// 使用 k8s 部署的 mysql
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	if err != nil {
		panic("failed to connect database")
	}

	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}

func initUser(db *gorm.DB, rdb redis.Cmdable) *handler.UserHandler {
	udao := dao.NewUserDao(db)
	uc := cache.NewUserCache(rdb)
	repo := repository.NewUserRepository(udao, uc)
	cc := cache.NewCodeCache(rdb)
	codeRepo := repository.NewCodeRepository(cc)
	smsSvc := memory.NewService()
	codeSvc := service.NewCodeService(codeRepo, smsSvc)
	svc := service.NewUserService(repo)
	u := handler.NewUserHandler(svc, codeSvc)
	return u
}

func initWebServer() *gin.Engine {
	server := gin.Default()

	server.Use(func(ctx *gin.Context) {
		fmt.Println("这是第一个Middleware")
		ctx.Next()
	})

	//redisClient := redis.NewClient(&redis.Options{
	//	// 使用 k8s 部署的redis
	//	Addr: "webook-redis:6379",
	//})
	//server.Use(ratelimit.NewBuilder(redisClient, time.Minute, 1000).Build())

	server.Use(cors.New(cors.Config{
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
	}))

	// store := cookie.NewStore([]byte("secret"))
	// 这是基于内存的实现，第一个参数为 authentication key ，最好为32位或者64位
	// 第二个参数为 encryption key
	// store := memstore.NewStore([]byte("Oh8wjuMwrYa#$&LN0c!6dmI5K6osZzvG"), []byte("oBSFwd5HKOSu86f7Q@AlmdRkkp@PCM*^"))

	// 第一个参数是最大空闲链接数量
	// 第二个就是 TCP，你不太可能用 udp
	// 第三个、四个 就是连接信息和密码
	// 第五个是 authentication key，指的是身份认证
	// 第六个是 encryption key，指的是数据加密，这两者加上权限控制，就是信息安全的三个核心概念

	//store, err := sessRedis.NewStore(16, "tcp", config.Config.Redis.Addr, "", []byte("Oh8wjuMwrYa#$&LN0c!6dmI5K6osZzvG"), []byte("oBSFwd5HKOSu86f7Q@AlmdRkkp@PCM*^"))
	//
	//if err != nil {
	//	panic(err)
	//}
	//server.Use(sessions.Sessions("mysession", store))

	// server.Use(middleware.NewLoginMiddlewareBuilder().
	// 	IgnorePaths("/users/login").
	// 	IgnorePaths("/users/signup").
	// 	Build())

	server.Use(middleware.NewLoginJWTMiddlewareBuilder().
		IgnorePaths("/users/login").
		IgnorePaths("/users/signup").
		IgnorePaths("/users/login_sms/code/send").
		IgnorePaths("/users/login_sms/code/verify").
		Build())

	return server
}
