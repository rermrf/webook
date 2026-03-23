package ioc

import (
	"github.com/spf13/viper"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	openapiv1 "webook/api/proto/gen/openapi/v1"
	"webook/credit/api"
	"webook/credit/service"
	"webook/pkg/logger"
)

func InitOpenAPIGRPCClient(etcdCli *etcdv3.Client) openapiv1.OpenAPIServiceClient {
	type Config struct {
		Target string `yaml:"target"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.openapi", &cfg)
	if err != nil {
		panic(err)
	}

	bd, err := resolver.NewBuilder(etcdCli)
	if err != nil {
		panic(err)
	}

	conn, err := grpc.NewClient(cfg.Target,
		grpc.WithResolvers(bd),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}

	return openapiv1.NewOpenAPIServiceClient(conn)
}

func InitCreditAPIHandler(
	openapiCli openapiv1.OpenAPIServiceClient,
	creditSvc service.CreditService,
	l logger.LoggerV1,
) *api.Handler {
	return api.NewHandler(openapiCli, creditSvc, l)
}
