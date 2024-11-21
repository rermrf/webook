package ioc

import (
	"github.com/spf13/viper"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	userv1 "webook/api/proto/gen/user/v1"
)

func InitUserGRPCClient(client *etcdv3.Client) userv1.UserServiceClient {
	type Config struct {
		Secure bool `yaml:"secure"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.user", &cfg)
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

	cc, err := grpc.NewClient("etcd:///service/user", opts...)
	if err != nil {
		panic(err)
	}
	return userv1.NewUserServiceClient(cc)
}
