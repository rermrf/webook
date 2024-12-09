package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	initViper()

	app := Init()
	for _, consumer := range app.Consumers {
		err := consumer.Start()
		if err != nil {
			panic(err)
		}
	}
	err := app.Server.Serve()
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
