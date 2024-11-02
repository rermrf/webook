// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

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

// Injectors from wire.go:

func InitArticleGRPCServer() *grpcx.Server {
	loggerV1 := ioc.InitLogger()
	db := ioc.InitDB(loggerV1)
	articleDao := dao.NewGormArticleDao(db)
	cmdable := ioc.InitRedis()
	articleCache := cache.NewRedisArticleCache(cmdable)
	articleRepository := repository.NewArticleRepository(articleDao, articleCache, loggerV1)
	syncProducer := ioc.InitProducer()
	producer := events.NewKafkaProducer(syncProducer)
	articleService := service.NewArticleService(articleRepository, loggerV1, producer)
	articleGRPCServer := grpc.NewArticleGRPCServer(articleService)
	server := ioc.InitGRPCServer(articleGRPCServer)
	return server
}

// wire.go:

var thirdPartySet = wire.NewSet(ioc.InitDB, ioc.InitLogger, ioc.InitRedis, ioc.InitProducer)

var articleSet = wire.NewSet(grpc.NewArticleGRPCServer, service.NewArticleService, repository.NewArticleRepository, dao.NewGormArticleDao, cache.NewRedisArticleCache)
