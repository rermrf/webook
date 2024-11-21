//go:build wireinject

package startup

import (
	"github.com/google/wire"
	pmtv1 "webook/api/proto/gen/payment/v1"
	"webook/reward/repository"
	"webook/reward/repository/cache"
	"webook/reward/repository/dao"
	"webook/reward/service"
)

var thirdPartySet = wire.NewSet(InitDB, InitLog, InitRedis)

func InitWechatNativeSvc(client pmtv1.WechatPaymentServiceClient) service.RewardService {
	wire.Build(
		thirdPartySet,
		service.NewWechatNativeRewardService,
		repository.NewRewardRepository,
		dao.NewRewardGORMDAO,
		cache.NewRewardRedisCache,
	)
	return &service.WechatNativeRewardService{}
}
