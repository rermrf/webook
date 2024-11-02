package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	initViper()

	server := InitUserGRPCServer()
	err := server.Serve()
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
