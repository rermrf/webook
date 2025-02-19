package ioc

import (
	"github.com/spf13/viper"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	commentv1 "webook/api/proto/gen/comment/v1"
)

func InitCommentGRPCClient(client *etcdv3.Client) commentv1.CommentServiceClient {
	type Config struct {
		Secure bool `yaml:"secure"`
	}
	var config Config
	err := viper.UnmarshalKey("grpc.client.comment", &config)
	if err != nil {
		panic(err)
	}

	db, err := resolver.NewBuilder(client)
	if err != nil {
		panic(err)
	}

	opts := []grpc.DialOption{grpc.WithResolvers(db)}
	if config.Secure {
		//opts = append(opts, grpc.WithInsecure())
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	cc, err := grpc.NewClient("etcd:///service/comment", opts...)
	if err != nil {
		panic(err)
	}
	return commentv1.NewCommentServiceClient(cc)
}
