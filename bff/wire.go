//go:build wireinject

package main

import (
	article3 "webook/article/events"
	bffevents "webook/bff/events"
	"webook/bff/handler"
	ijwt "webook/bff/handler/jwt"
	"webook/bff/ioc"
	"webook/interactive/events"
	repository2 "webook/interactive/repository"
	cache2 "webook/interactive/repository/cache"
	dao2 "webook/interactive/repository/dao"
	service2 "webook/interactive/service"

	"github.com/google/wire"
)

// User 相关依赖
var UserSet = wire.NewSet(
	handler.NewUserHandler,
)

// Gorm 文章相关依赖
var GormArticleSet = wire.NewSet(
	handler.NewArticleHandler,
)

// FollowSet 关注相关依赖
var FollowSet = wire.NewSet(
	handler.NewFollowHandler,
)

var SearchSet = wire.NewSet(
	handler.NewSearchHandler,
)

// TagSet 标签相关依赖
var TagSet = wire.NewSet(
	handler.NewTagHandler,
	ioc.InitTagGRPCClient,
)

// FeedSet Feed 相关依赖
var FeedSet = wire.NewSet(
	handler.NewFeedHandler,
	ioc.InitFeedGRPCClient,
)

// RankingSet 排行榜相关依赖
var RankingSet = wire.NewSet(
	handler.NewRankingHandler,
)

// NotificationSet 通知相关依赖
var NotificationSet = wire.NewSet(
	handler.NewNotificationHandler,
	ioc.InitNotificationGRPCClient,
	ioc.InitSSEHub,
)

// NotificationProducerSet 通知事件生产者
var NotificationProducerSet = wire.NewSet(
	bffevents.NewSaramaNotificationProducer,
)

var ThirdPartySet = wire.NewSet(
	ioc.InitRedis,
	ioc.InitDB,
	ioc.InitLogger,
	ijwt.NewRedisJWTHandler,
	ioc.InitEtcd,
)

var InteractiveSet = wire.NewSet(
	service2.NewInteractiveService,
	repository2.NewCachedInteractiveRepository,
	dao2.NewGORMInteractiveDao,
	cache2.NewRedisInteractiveCache,
	events.NewInteractiveReadBatchConsumer,
)

var OAuth2Set = wire.NewSet(
	handler.NewOAuth2WechatHandler,
)

var KafkaSet = wire.NewSet(
	ioc.InitKafka,
	ioc.NewConsumer,
	ioc.NewSyncProducer,
	article3.NewKafkaProducer,
)

var grpcClientSet = wire.NewSet(
	ioc.InitIntrGRPCClientV2,
	ioc.InitUserGRPCClient,
	ioc.InitArticleGRPCClientV1,
	ioc.InitSMSGRPCClient,
	ioc.InitCodeGRPCClient,
	ioc.InitRankingGRPCClient,
	ioc.InitOAuth2GRPCClient,
	ioc.InitRewardGRPCClient,
	ioc.InitFollowGRPCClient,
	ioc.InitCommentGRPCClient,
	ioc.InitSearchGRPCClient,
	ioc.InitCreditGRPCClient,
)

var CreditSet = wire.NewSet(
	handler.NewCreditHandler,
)

// IMSet IM 私信相关依赖
var IMSet = wire.NewSet(
	handler.NewIMHandler,
	ioc.InitIMGRPCClient,
	ioc.InitIMHub,
)

// HistorySet 浏览历史相关依赖
var HistorySet = wire.NewSet(
	handler.NewHistoryHandler,
	ioc.InitHistoryGRPCClient,
)

func InitApp() *App {
	wire.Build(
		// 中间件，路由等？
		//gin.Default,
		ioc.InitGin,
		ioc.InitMiddlewares,
		UserSet,
		FollowSet,
		SearchSet,
		NotificationSet,
		NotificationProducerSet,
		CreditSet,
		IMSet,
		HistorySet,
		TagSet,
		FeedSet,
		RankingSet,

		//CodeSet,
		ThirdPartySet,
		OAuth2Set,
		InteractiveSet,
		grpcClientSet,
		KafkaSet,
		GormArticleSet,
		// 组装我这个结构体的所有字段

		ioc.InitRankingJob,
		ioc.InitJob,
		ioc.InitRLockClient,

		wire.Struct(new(App), "*"),
	)
	return new(App)
}
