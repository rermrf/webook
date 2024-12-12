package test

import (
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
	"testing"
	feedv1 "webook/api/proto/gen/feed"
	followMocks "webook/api/proto/gen/follow/v1/mocks"
	"webook/feed/grpc"
	"webook/feed/ioc"
	"webook/feed/repository"
	"webook/feed/repository/cache"
	"webook/feed/repository/dao"
	"webook/feed/service"
)

func InitGrpcServer(t *testing.T) (feedv1.FeedSvcServer, *followMocks.MockFollowServiceClient, *gorm.DB) {
	logger := ioc.InitLogger()
	db := ioc.InitDB(logger)
	feedPullEventDao := dao.NewFeedPullEventDao(db)
	feedPushEventDao := dao.NewFeedPushEventDao(db)
	cmdable := ioc.InitRedis()
	feedEventCache := cache.NewFeedEventCache(cmdable)
	feedEventRepo := repository.NewFeedEventRepository(feedPullEventDao, feedPushEventDao, feedEventCache)
	mockCtrl := gomock.NewController(t)
	followClient := followMocks.NewMockFollowServiceClient(mockCtrl)
	v := ioc.RegisterHandler(feedEventRepo, followClient)
	feedService := service.NewFeedService(feedEventRepo, v, followClient)
	feedEventGrpcSvc := grpc.NewFeedEventGrpcServer(feedService)
	return feedEventGrpcSvc, followClient, db
}
