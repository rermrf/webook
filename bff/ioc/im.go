package ioc

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	imv1 "webook/api/proto/gen/im/v1"
	"webook/bff/handler/ws"
	"webook/pkg/logger"
)

func InitIMGRPCClient(client *etcdv3.Client) imv1.IMServiceClient {
	type Config struct {
		Secure bool `yaml:"secure"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.im", &cfg)
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

	cc, err := grpc.NewClient("etcd:///service/im", opts...)
	if err != nil {
		panic(err)
	}
	return imv1.NewIMServiceClient(cc)
}

// InitIMHub 初始化 IM WebSocket Hub
func InitIMHub(redisClient redis.Cmdable, imSvc imv1.IMServiceClient, l logger.LoggerV1) *ws.IMHub {
	hub := ws.NewIMHub(redisClient, imSvc, l)
	go hub.Run(context.Background())
	return hub
}
