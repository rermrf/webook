package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
)

func main() {

	initViper()

	app := InitApp()

	// 使用消费者
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}

	go func() {
		err := app.webAdmin.Start()
		log.Println(err)
	}()

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
