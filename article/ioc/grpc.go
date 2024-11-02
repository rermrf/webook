package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"log"
	igrpc "webook/article/grpc"
	"webook/pkg/grpcx"
)

func InitGRPCServer(artServer *igrpc.ArticleGRPCServer) *grpcx.Server {
	type Config struct {
		Addr string `yaml:"addr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	server := grpc.NewServer()
	artServer.Register(server)
	log.Println("article server will work on: " + cfg.Addr)
	return &grpcx.Server{
		Server: server,
		Addr:   cfg.Addr,
	}
}
