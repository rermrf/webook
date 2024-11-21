package ioc

import (
	"github.com/spf13/viper"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	codev1 "webook/api/proto/gen/code/v1"
)

func InitCodeGRPCClient(client *etcdv3.Client) codev1.CodeServiceClient {
	type Config struct {
		Secure bool `yaml:"secure"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.code", &cfg)
	if err != nil {
		panic(err)
	}

	bd, err := resolver.NewBuilder(client)
	if err != nil {
		panic(err)
	}

	opts := []grpc.DialOption{grpc.WithResolvers(bd)}
	if cfg.Secure {
		// 加载证书之类的东西
		// 启用 HTTPS
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	cc, err := grpc.NewClient("etcd:///service/code", opts...)
	if err != nil {
		panic(err)
	}
	return codev1.NewCodeServiceClient(cc)
}
