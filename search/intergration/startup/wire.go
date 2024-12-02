//go:build wireinject

package startup

import (
	"github.com/google/wire"
	"webook/search/grpc"
	"webook/search/repository"
	"webook/search/repository/dao"
	"webook/search/service"
)

var thirdPartySet = wire.NewSet(
	InitLog,
	InitESClient,
)

var serviceSet = wire.NewSet(
	service.NewSearchService,
	service.NewSyncService,
	repository.NewAnyRepository,
	repository.NewArticleRepository,
	repository.NewUserRepository,
	dao.NewUserESDao,
	dao.NewTagESDao,
	dao.NewArticleESDao,
	dao.NewAnyESDao,
)

func InitSearchServer() *grpc.SearchServiceServer {
	wire.Build(
		thirdPartySet,
		serviceSet,
		grpc.NewSearchServiceServer,
	)
	return new(grpc.SearchServiceServer)
}

func InitSyncServer() *grpc.SyncServiceServer {
	wire.Build(
		thirdPartySet,
		serviceSet,
		grpc.NewSyncServiceServer,
	)
	return new(grpc.SyncServiceServer)
}
