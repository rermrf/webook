package ioc

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	intrv1 "webook/api/proto/gen/intr/v1"
	"webook/interactive/service"
	"webook/internal/client"
)

func InitEtcd() *etcdv3.Client {
	type Config struct {
		Addr string `yaml:"addr"`
	}
	var cfg Config
	if err := viper.UnmarshalKey("etcd", &cfg); err != nil {
		panic(err)
	}
	cli, err := etcdv3.NewFromURL(cfg.Addr)
	if err != nil {
		panic(err)
	}
	return cli
}

func InitIntrGRPCClientV2(client *etcdv3.Client) intrv1.InteractiveServiceClient {
	type Config struct {
		Secure bool `yaml:"secure"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.intr", &cfg)
	if err != nil {
		panic(err)
	}

	bd, err := resolver.NewBuilder(client)
	if err != nil {
		panic(err)
	}

	opts := []grpc.DialOption{grpc.WithResolvers(bd), grpc.WithInsecure()}
	if cfg.Secure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	opts = append(opts, grpc.WithResolvers(bd))
	conn, err := grpc.NewClient("etcd:///service/interactive", opts...)
	if err != nil {
		panic(err)
	}
	return intrv1.NewInteractiveServiceClient(conn)

}

// InitIntrGRPCClientV1 真正的 gRPC 客户端
func InitIntrGRPCClientV1() intrv1.InteractiveServiceClient {
	type Config struct {
		Addr     string `yaml:"addr"`
		Secure   bool   `yaml:"secure"`
		EtcdAddr string `yaml:"etcdAddr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.intr", &cfg)
	if err != nil {
		panic(err)
	}
	cli, err := etcdv3.NewFromURL(cfg.EtcdAddr)
	if err != nil {
		panic(err)
	}
	etcdResolver, err := resolver.NewBuilder(cli)
	if err != nil {
		panic(err)
	}
	opts := []grpc.DialOption{grpc.WithResolvers(etcdResolver), grpc.WithInsecure()}
	if cfg.Secure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	conn, err := grpc.Dial(cfg.Addr, opts...)
	if err != nil {
		panic(err)
	}
	return intrv1.NewInteractiveServiceClient(conn)
}

// InitIntrGRPCClient 流量控制的客户端
func InitIntrGRPCClient(svc service.InteractiveService) intrv1.InteractiveServiceClient {
	type Config struct {
		Addr      string `yaml:"addr"`
		Secure    bool   `yaml:"secure"`
		Threshold int32  `yaml:"threshold"`
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
	remote := intrv1.NewInteractiveServiceClient(cc)
	local := client.NewInteractiveServiceAdapter(svc)
	res := client.NewGreyScaleInteractiveServiceClient(remote, local)
	res.UpdateThreshold(cfg.Threshold)
	// 监听配置文件变更
	viper.OnConfigChange(func(e fsnotify.Event) {
		var cfg Config
		err := viper.UnmarshalKey("grpc.client.intr", &cfg)
		if err != nil {
			// 可以记录日志
		}
		res.UpdateThreshold(cfg.Threshold)
	})
	return res
}
