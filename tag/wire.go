//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/pkg/grpcx"
	"webook/tag/grpc"
	"webook/tag/ioc"
	"webook/tag/repository/cache"
	"webook/tag/repository/dao"
	"webook/tag/service"
)

var thirdPartySet = wire.NewSet(
	ioc.InitDB,
	ioc.InitRedis,
	ioc.InitLogger,
	ioc.InitEtcd,
	ioc.InitKafka,
)

func InitTagGRPCServer() *grpcx.Server {
	wire.Build(
		thirdPartySet,
		ioc.InitGRPCServer,
		ioc.InitProducer,
		grpc.NewTagServiceServer,
		service.NewTagService,
		ioc.InitRepository,
		dao.NewGORMTagDao,
		cache.NewRedisTagCache,
	)
	return new(grpcx.Server)
}
