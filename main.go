package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"go.uber.org/zap"
	"net/http"
	"time"
	"webook/config"
	"webook/internal/ioc"
)

func main() {
	//db := initDB()
	//rdb := initRDB()
	//server := initWebServer()
	//u := initUser(db, rdb)
	//u.RegisterRoutes(server)

	initViperV1()
	initLogger()

	closeFunc := ioc.InitOTEL()

	initPrometheus()
	// etcdctl --endpoints=127.0.0.1:12379 put /webook "$(<./config/dev.yaml)"
	//initViperReomte()

	app := InitWebServer()

	// kafka 消费
	for _, c := range app.Consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}

	app.cron.Start()

	server := app.Server
	err := server.Run(config.Config.Server.HTTPPort)
	if err != nil {
		return
	}

	// 一分钟内关完
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	closeFunc(ctx)

	cx := app.cron.Stop()
	// 想办法 close
	// 可以考虑超时强制退出，防止有些任务执行特别长的时间
	tm := time.NewTimer(time.Minute * 10)
	select {
	case <-tm.C:
	case <-cx.Done():
	}
}

func initPrometheus() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8082", nil)
	}()
}

func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	// 如果你不 replace，直接用 zap.L()，啥都打印不了
	zap.ReplaceGlobals(logger)
	zap.L().Info("")
}

func initViperReomteWatch() {
	viper.SetConfigType("yaml")
	err := viper.AddRemoteProvider("etcd3", "http://localhost:12379", "/webook")
	if err != nil {
		panic(err)
	}
	err = viper.WatchRemoteConfig()
	if err != nil {
		panic(err)
	}
	// 远程并不能监听
	//viper.OnConfigChange(func(in fsnotify.Event) {
	//	fmt.Println("config file changed:", in.Name)
	//})
	err = viper.ReadRemoteConfig()
	if err != nil {
		panic(err)
	}
}

// 使用配置中心的方式
func initViperReomte() {
	viper.SetConfigType("yaml")
	err := viper.AddRemoteProvider("etcd3", "http://localhost:12379", "/webook")
	if err != nil {
		panic(err)
	}
	err = viper.ReadRemoteConfig()
	if err != nil {
		panic(err)
	}
}

func initViperV2Watch() {
	file := pflag.String("config", "./config/dev.yaml", "指定配置文件路径")
	pflag.Parse()
	viper.SetConfigFile(*file)
	// 实时监听配置变更
	viper.WatchConfig()
	// 当发生改变时，执行里面的代码快，用来替换之前的对象
	viper.OnConfigChange(func(in fsnotify.Event) {
		// 并没有告诉你，变更前和变更后的数据，需要自己手动读取
		fmt.Println(in.Name, in.Op)
		fmt.Println(viper.GetString("db.mysql.dsn"))
	})

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

}

func initViperV1() {
	file := pflag.String("config", "./config/dev.yaml", "指定配置文件路径")
	pflag.Parse()
	viper.SetConfigFile(*file)

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	//viper.SetConfigFile("./config/dev.yaml")
	//err := viper.ReadInConfig()
	//if err != nil {
	//	panic(fmt.Errorf("Fatal error config file: %s \n", err))
	//}
}

func initViperReader() {
	viper.SetConfigType("yaml")
	cfg := `
	db.mysql:
  dsn: "root:root@tcp(localhost:13306)/webook?charset=utf8mb4&parseTime=True&loc=Local"

redis:
  addr: "localhost:6379"
`
	err := viper.ReadConfig(bytes.NewReader([]byte(cfg)))
	if err != nil {
		panic(err)
	}
}

func initViper() {
	// viper 设置默认值
	viper.SetDefault("db.mysql.dsn", "root:root@tcp(localhost:3306)/webook?charset=utf8mb4&parseTime=True&loc=Local")
	// 配置文件的名字，不包含文件名的扩展名
	viper.SetConfigName("dev")
	// 告诉 viper 使用 yaml 格式
	viper.SetConfigType("yaml")
	// 设置配置文件路径，可以有多个
	viper.AddConfigPath("./config")
	// 读取配置文件到 viper 中
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}

//func initWebServer() *gin.Engine {
//	server := gin.Default()
//
//	server.Use(func(ctx *gin.Context) {
//		fmt.Println("这是第一个Middleware")
//		ctx.Next()
//	})
//
//	//redisClient := redis.NewClient(&redis.Options{
//	//	// 使用 k8s 部署的redis
//	//	Addr: "webook-redis:6379",
//	//})
//	//server.Use(ratelimit.NewBuilder(redisClient, time.Minute, 1000).Build())
//
//	server.Use(cors.New(cors.Config{
//		// AllowOrigins: []string{"http://localhost:3000"},
//		AllowMethods: cors.DefaultConfig().AllowMethods,
//		AllowHeaders: []string{"Content-Type", "Authorization"},
//		// 不加这个，前端拿不到
//		ExposeHeaders:    []string{"x-jwt-token"},
//		AllowCredentials: true, // 是否允许发送Cookie，默认false
//		AllowOriginFunc: func(origin string) bool {
//			if strings.HasPrefix(origin, "http://localhost") {
//				// 开发环境
//				return true
//			}
//			return strings.Contains(origin, "yourcompany.com") // 允许公司域名访问
//		},
//		MaxAge: 12 * time.Hour,
//	}))
//
//	// store := cookie.NewStore([]byte("secret"))
//	// 这是基于内存的实现，第一个参数为 authentication key ，最好为32位或者64位
//	// 第二个参数为 encryption key
//	// store := memstore.NewStore([]byte("Oh8wjuMwrYa#$&LN0c!6dmI5K6osZzvG"), []byte("oBSFwd5HKOSu86f7Q@AlmdRkkp@PCM*^"))
//
//	// 第一个参数是最大空闲链接数量
//	// 第二个就是 TCP，你不太可能用 udp
//	// 第三个、四个 就是连接信息和密码
//	// 第五个是 authentication key，指的是身份认证
//	// 第六个是 encryption key，指的是数据加密，这两者加上权限控制，就是信息安全的三个核心概念
//
//	//store, err := sessRedis.NewStore(16, "tcp", config.Config.Redis.Addr, "", []byte("Oh8wjuMwrYa#$&LN0c!6dmI5K6osZzvG"), []byte("oBSFwd5HKOSu86f7Q@AlmdRkkp@PCM*^"))
//	//
//	//if err != nil {
//	//	panic(err)
//	//}
//	//server.Use(sessions.Sessions("mysession", store))
//
//	// server.Use(middleware.NewLoginMiddlewareBuilder().
//	// 	IgnorePaths("/users/login").
//	// 	IgnorePaths("/users/signup").
//	// 	Build())
//
//	server.Use(middleware.NewLoginJWTMiddlewareBuilder().
//		IgnorePaths("/users/login").
//		IgnorePaths("/users/signup").
//		IgnorePaths("/users/login_sms/code/send").
//		IgnorePaths("/users/login_sms/code/verify").
//		IgnorePaths("/users/login_sms").
//		Build())
//
//	return server
//}
