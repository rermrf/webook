//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/pkg/grpcx"
	igrpc "webook/ranking/grpc"
	"webook/ranking/ioc"
	"webook/ranking/repository"
	"webook/ranking/repository/cache"
	"webook/ranking/service"
)

var grpcClientSet = wire.NewSet(
	ioc.InitArticleGRPCClient,
	ioc.InitIntrGRPCClient,
)

func InitRankingGRPCServer() *grpcx.Server {
	wire.Build(
		ioc.InitRedis,
		//ioc.InitLogger,
		cache.NewRankingLocalCache,
		cache.NewRankingRedisCache,
		repository.NewCachedRankingRepository,
		service.NewBatchRankingService,
		igrpc.NewRankingServiceServer,
		ioc.InitGRPCServer,

		grpcClientSet,
	)
	return new(grpcx.Server)
}
