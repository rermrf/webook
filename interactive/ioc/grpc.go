package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"log"
	grpc2 "webook/interactive/grpc"
	"webook/pkg/grpcx"
)

func InitGRPCServer(intrServer *grpc2.InteractiveServiceServer) *grpcx.Server {
	type Config struct {
		Addr string `yaml:"addr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}

	server := grpc.NewServer()
	intrServer.Register(server)
	log.Println(cfg.Addr)

	return &grpcx.Server{
		Server: server,
		Addr:   cfg.Addr,
	}
}
