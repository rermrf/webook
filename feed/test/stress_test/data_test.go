package stress_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"math/rand"
	"net/http"
	"testing"
	"time"
	feedv1 "webook/api/proto/gen/feed"
	followv1 "webook/api/proto/gen/follow/v1"
	followMocks "webook/api/proto/gen/follow/v1/mocks"
	"webook/feed/ioc"
	"webook/feed/repository"
	"webook/feed/repository/cache"
	"webook/feed/repository/dao"
	"webook/feed/service"
	"webook/feed/test"
	"webook/feed/test/stress_test/web"
)

// 生成拉事件
func generatePullEvent(mockFollowClient *followMocks.MockFollowServiceClient, id int64) test.ArticleEvent {
	mockFollowClient.EXPECT().GetFollowStatic(gomock.Any(), &followv1.GetFollowStaticRequest{
		Followee: id,
	}).Return(&followv1.GetFollowStaticResponse{
		FollowStatic: &followv1.FollowStatic{
			Followees: 1000,
		},
	}, nil)
	return test.ArticleEvent{
		Uid:   fmt.Sprintf("%d", id),
		Aid:   fmt.Sprintf("%d", id),
		Title: fmt.Sprintf("%d发布了文章", id),
	}
}

// 生成推事件
func generatePushEvent(mockFollowClient *followMocks.MockFollowServiceClient, id int64, i int64) test.ArticleEvent {
	// 生成几个推事件都包含id i
	mockFollowClient.EXPECT().GetFollowStatic(gomock.Any(), &followv1.GetFollowStaticRequest{
		Followee: id + i,
	}).Return(&followv1.GetFollowStaticResponse{
		FollowStatic: &followv1.FollowStatic{
			Followees: 2,
		},
	}, nil)
	mockFollowClient.EXPECT().GetFollower(gomock.Any(), &followv1.GetFollowerRequest{
		Followee: id + i,
	}).Return(&followv1.GetFollowerResponse{
		FollowRelations: []*followv1.FollowRelation{
			{
				Id:       time.Now().UnixNano(),
				Followee: id + i,
				Follower: id,
			},
			{
				Id:       time.Now().UnixNano(),
				Follower: id + i + 1,
				Followee: id,
			},
		},
	}, nil)
	return test.ArticleEvent{
		Uid:   fmt.Sprintf("%d", id+i),
		Aid:   fmt.Sprintf("%d", time.Now().UnixNano()),
		Title: fmt.Sprintf("%d发布了文章", id+i),
	}
}

// 生成数据
func Test_AddFeed(t *testing.T) {
	server, followCliet, _ := test.InitGrpcServer(t)
	// 生成拉事件的压力测试数据
	for i := 2; i < 100000; i++ {
		event := generatePullEvent(followCliet, int64(i))
		ext, _ := json.Marshal(event)
		_, err := server.CreateFeedEvent(context.Background(), &feedv1.CreateFeedEventRequest{
			FeedEvent: &feedv1.FeedEvent{
				Type:    service.ArticleEventName,
				Content: string(ext),
			},
		})
		require.NoError(t, err)
	}

	// 生成拉事件的压力测试数据
	for i := 0; i < 100000; i++ {
		event := generatePushEvent(followCliet, int64(400001), int64(i))
		ext, _ := json.Marshal(event)
		_, err := server.CreateFeedEvent(context.Background(), &feedv1.CreateFeedEventRequest{
			FeedEvent: &feedv1.FeedEvent{
				Type:    service.ArticleEventName,
				Content: string(ext),
			},
		})
		require.NoError(t, err)
	}
}

// follow 用的是本地 mock，所以这个测试排除了 follow 的错误
func Test_Feed(t *testing.T) {
	viper.SetConfigFile("./config.yaml")
	viper.WatchConfig()
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	loggerV1 := ioc.InitLogger()
	db := ioc.InitDB(loggerV1)
	feedPullEventDao := dao.NewFeedPullEventDao(db)
	feedPushEventDao := dao.NewFeedPushEventDao(db)
	cmdable := ioc.InitRedis()
	feedEventCache := cache.NewFeedEventCache(cmdable)
	feedEventRepo := repository.NewFeedEventRepository(feedPullEventDao, feedPushEventDao, feedEventCache)
	mockCtrl := gomock.NewController(t)
	// 不想用 mock，就用真实的 follow rpc client
	// 想要模拟降级，在 follow 加上降级的逻辑
	followClient := followMocks.NewMockFollowServiceClient(mockCtrl)
	v := ioc.RegisterHandler(feedEventRepo, followClient)
	feedService := service.NewFeedService(feedEventRepo, v, followClient)
	engine := gin.Default()
	handler := web.NewFeedHandler(feedService)
	handler.RegisterRoutes(engine)
	// 设置mock数据
	// 设置关注列表的测试数据
	followClient.EXPECT().GetFollowee(gomock.Any(), gomock.Any()).Return(&followv1.GetFolloweeResponse{
		FollowRelations: getFollowRelation(1),
	}, nil).AnyTimes()
	// 设置粉丝列表的测试数据
	// 扩散百人
	followClient.EXPECT().GetFollowStatic(gomock.Any(), &followv1.GetFollowStaticRequest{
		Followee: 4,
	}).Return(&followv1.GetFollowStaticResponse{
		FollowStatic: &followv1.FollowStatic{
			Followees: 800,
		},
	}, nil).AnyTimes()
	followClient.EXPECT().GetFollower(gomock.Any(), &followv1.GetFollowerRequest{
		Followee: 4,
	}).Return(&followv1.GetFollowerResponse{
		FollowRelations: getFollowrRelation(4, 800),
	}, nil).AnyTimes()
	// 扩散千人
	followClient.EXPECT().GetFollowStatic(gomock.Any(), &followv1.GetFollowerRequest{
		Followee: 5,
	}).Return(&followv1.GetFollowStaticResponse{
		FollowStatic: &followv1.FollowStatic{
			Followees: 5000,
		},
	}, nil).AnyTimes()
	followClient.EXPECT().GetFollower(gomock.Any(), &followv1.GetFollowerRequest{
		Followee: 5,
	}).Return(&followv1.GetFollowerResponse{
		FollowRelations: getFollowrRelation(5, 5000),
	}, nil).AnyTimes()
	// 扩散万人
	followClient.EXPECT().GetFollowStatic(gomock.Any(), &followv1.GetFollowerRequest{
		Followee: 6,
	}).Return(&followv1.GetFollowStaticResponse{
		FollowStatic: &followv1.FollowStatic{
			Followees: 50000,
		},
	}, nil).AnyTimes()
	followClient.EXPECT().GetFollower(gomock.Any(), &followv1.GetFollowerRequest{
		Followee: 6,
	}).Return(&followv1.GetFollowerResponse{
		FollowRelations: getFollowrRelation(6, 10000),
	}, nil).AnyTimes()
	engine.Run("127.0.0.1:8088")
}

func getFollowRelation(id int64) []*followv1.FollowRelation {
	relations := make([]*followv1.FollowRelation, 0, 100001)
	random := rand.Intn(200) + 300
	for i := random - 200; i < random; i++ {
		relations = append(relations, &followv1.FollowRelation{
			Follower: id,
			Followee: int64(i),
		})
	}
	return relations
}

func getFollowrRelation(id int64, number int) []*followv1.FollowRelation {
	relations := make([]*followv1.FollowRelation, 0, 100001)
	for i := 0; i < number+1; i++ {
		relations = append(relations, &followv1.FollowRelation{
			Follower: id,
			Followee: int64(i),
		})
	}
	return relations
}

func TestHello(t *testing.T) {
	server := gin.Default()
	server.POST("/hello", func(ctx *gin.Context) {
		var u User
		ctx.Bind(&u)
		r := rand.Int31n(1000)
		time.Sleep(time.Millisecond * time.Duration(r))
		// 这里我们模拟一下错误
		// 模拟 10% 的请求失败
		if r%100 < 10 {
			ctx.String(http.StatusInternalServerError, "系统错误")
		} else {
			ctx.String(http.StatusOK, "hello %s", u.Name)
		}
	})
	server.Run(":8080")
}

type User struct {
	Name string `json:"name"`
}
