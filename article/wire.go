//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/article/events"
	"webook/article/grpc"
	"webook/article/ioc"
	"webook/article/repository"
	"webook/article/repository/cache"
	"webook/article/repository/dao"
	"webook/article/service"
	"webook/pkg/grpcx"
)

var thirdPartySet = wire.NewSet(
	ioc.InitDB,
	ioc.InitLogger,
	ioc.InitRedis,
	ioc.InitProducer,
)

var articleSet = wire.NewSet(
	grpc.NewArticleGRPCServer,
	service.NewArticleService,
	repository.NewArticleRepository,
	dao.NewGormArticleDao,
	cache.NewRedisArticleCache,
)

func InitArticleGRPCServer() *grpcx.Server {
	wire.Build(
		thirdPartySet,
		articleSet,
		ioc.InitGRPCServer,
		events.NewKafkaProducer,
	)
	return new(grpcx.Server)
}
