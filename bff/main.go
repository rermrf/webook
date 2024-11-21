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

	initViperV2Watch()
	//initLogger()

	closeFunc := ioc.InitOTEL()

	initPrometheus()
	// etcdctl --endpoints=127.0.0.1:12379 put /webook "$(<./config/dev.yaml)"
	//initViperReomte()

	app := InitApp()

	// kafka 消费
	for _, c := range app.Consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}

	// HTTP 方式使用迁移的方法
	//go func() {
	//	scheduler := scheduler2.NewScheduler()
	//	scheduler.RegisterRoutes(app.webAdmin.group("/migrator"))
	//	app.webAdmin.Run(":8081")
	//}()

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

// 不使用依赖注入的时候使用
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
