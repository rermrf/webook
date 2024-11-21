package ioc

import (
	"github.com/spf13/viper"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	oauth2v1 "webook/api/proto/gen/oauth2/v1"
)

func InitOAuth2GRPCClient(client *etcdv3.Client) oauth2v1.Oauth2ServiceClient {
	type Config struct {
		Secure bool `yaml:"secure"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.oauth2", &cfg)
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

	cc, err := grpc.NewClient("etcd:///service/oauth2", opts...)
	if err != nil {
		panic(err)
	}
	return oauth2v1.NewOauth2ServiceClient(cc)
}
