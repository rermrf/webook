//go:build wireinject

package startup

import (
	"github.com/google/wire"
	"webook/account/grpc"
	"webook/account/repository"
	"webook/account/repository/dao"
	"webook/account/service"
)

func InitAccountServiceServer() *grpc.AccountServiceServer {
	wire.Build(
		InitDB,
		dao.NewAccountGORMDAO,
		repository.NewAccountRepository,
		service.NewAccountService,
		grpc.NewAccountServiceServer,
	)
	return &grpc.AccountServiceServer{}
}
