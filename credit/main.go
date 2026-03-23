package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	initViper()

	app := InitApp()

	// 启动 Kafka 消费者
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}

	// 启动 HTTP 服务（开放API）
	go func() {
		err := app.httpServer.Start()
		if err != nil {
			panic(err)
		}
	}()

	// 启动 gRPC 服务
	err := app.server.Serve()
	if err != nil {
		panic(err)
	}
}

func initViper() {
	file := pflag.String("config", "config/dev.yaml", "指定配置文件路径")
	pflag.Parse()
	viper.SetConfigFile(*file)

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}
