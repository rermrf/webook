package main

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	initViper()
	app := InitApp()

	// 启动 gRPC 服务
	err := app.server.Serve()
	if err != nil {
		panic(err)
	}
}

func initViper() {
	cfile := pflag.String("config", "config/dev.yaml", "配置文件路径")
	pflag.Parse()
	viper.SetConfigFile(*cfile)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}
