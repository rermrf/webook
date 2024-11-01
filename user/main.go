package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"net"
	userv1 "webook/api/proto/gen/user/v1"
)

func main() {
	initViper()

	server := grpc.NewServer()
	l, err := net.Listen("tcp", ":8091")
	if err != nil {
		panic(err)
	}
	userv1.RegisterUserServiceServer(server, InitUserGRPCServer())
	err = server.Serve(l)

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
