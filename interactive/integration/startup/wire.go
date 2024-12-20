//go:build wireinject

package startup

import (
	"github.com/google/wire"
	"webook/interactive/grpc"
	"webook/interactive/repository"
	"webook/interactive/repository/cache"
	"webook/interactive/repository/dao"
	"webook/interactive/service"
)

var thirdPartySet = wire.NewSet(
	InitDB,
	InitRedis,
	InitLog,
)

var interactiveSet = wire.NewSet(
	service.NewInteractiveService,
	repository.NewCachedInteractiveRepository,
	dao.NewGORMInteractiveDao,
	cache.NewRedisInteractiveCache,
)

func InitInteractiveService() service.InteractiveService {
	wire.Build(
		thirdPartySet,
		interactiveSet,
	)
	return service.NewInteractiveService(nil)
}

// 测试 server
func InitInteractiveGRPCServer() *grpc.InteractiveServiceServer {
	wire.Build(
		thirdPartySet,
		interactiveSet,
		grpc.NewInteractiveServiceServer,
	)
	return new(grpc.InteractiveServiceServer)
}
