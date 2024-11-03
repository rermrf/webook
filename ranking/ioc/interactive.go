package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	intrv1 "webook/api/proto/gen/intr/v1"
)

func InitIntrGRPCClient() intrv1.InteractiveServiceClient {
	type Config struct {
		Addr   string `yaml:"addr"`
		Secure bool   `yaml:"secure"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.intr", &cfg)
	if err != nil {
		panic(err)
	}
	var opts []grpc.DialOption
	if cfg.Secure {
		// 加载证书之类的东西
		// 启用 HTTPS
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	cc, err := grpc.NewClient(cfg.Addr, opts...)
	if err != nil {
		panic(err)
	}
	return intrv1.NewInteractiveServiceClient(cc)
}
