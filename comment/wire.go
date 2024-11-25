//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/comment/grpc"
	"webook/comment/ioc"
	"webook/comment/repository"
	"webook/comment/repository/dao"
	"webook/comment/service"
	"webook/pkg/grpcx"
)

var thirdPartySet = wire.NewSet(
	ioc.InitEtcd,
	ioc.InitLogger,
	ioc.InitDB,
)

func InitCommentGRPCServiceServer() *grpcx.Server {
	wire.Build(
		thirdPartySet,
		ioc.InitGRPCServer,
		grpc.NewCommentServiceServer,
		service.NewCommentService,
		repository.NewCommentRepository,
		dao.NewGORMCommentDAO,
	)
	return new(grpcx.Server)
}
