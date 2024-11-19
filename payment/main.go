package main

import (
	"fmt"
	"github.com/IBM/sarama"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"webook/pkg/ginx"
	"webook/pkg/grpcx"
)

type App struct {
	GRPCServer *grpcx.Server
	WebServer  *ginx.Server
	Consumer   []sarama.Consumer
}

func main() {
	initViper()
	app := InitApp()
	go func() {
		err := app.GRPCServer.Serve()
		if err != nil {
			panic(err)
		}
	}()
	err := app.WebServer.Start()
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
