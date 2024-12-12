package test

import (
	"context"
	"encoding/json"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
	feedv1 "webook/api/proto/gen/feed"
	followv1 "webook/api/proto/gen/follow/v1"
	followMocks "webook/api/proto/gen/follow/v1/mocks"
	"webook/feed/repository/dao"
	"webook/feed/service"
)

// 测试主流程，创建推事件，创建拉事件
type FeedTestSuite struct {
	suite.Suite
}

func (f *FeedTestSuite) SetupSuite() {
	//初始化配置文件
	viper.SetConfigFile("config.yaml")
	viper.WatchConfig()
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func (f *FeedTestSuite) Test_Feed() {
	// 初始化
	server, mockFollowClient, db := InitGrpcServer(f.T())
	defer func() {
		db.Table("feed_push_events").Where("id > ?", 0).Delete(&dao.FeedPushEvent{})
		db.Table("feed_pull_events").Where("id > ?", 0).Delete(&dao.FeedPullEvent{})
	}()
	// 设置 followmock 的值
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Minute)
	defer cancel()
	// 创建时间
	err := f.setupEvent(ctx, mockFollowClient, server)
	require.NoError(f.T(), err)
	// 创建 feed 流事件
	wantEvents := f.getFeedEvent(ctx, mockFollowClient, server)
	resp, err := server.FindFeedEvents(ctx, &feedv1.FindFeedEventsRequest{
		Uid:       1,
		Limit:     20,
		Timestamp: time.Now().Unix() + 3,
	})
	require.NoError(f.T(), err)
	assert.Equal(f.T(), len(wantEvents), len(resp.FeedEvents))
	checkerMap := map[string]EventCheck{
		service.ArticleEventName: ArticleEvent{},
		service.LikeEventName:    LikeEvent{},
		service.FollowEventName:  FollowEvent{},
	}
	for i := 0; i < len(wantEvents); i++ {
		wantEvent, actualEvent := wantEvents[i], resp.FeedEvents[i]
		checker := checkerMap[wantEvent.Type]
		wantContent, actualContent := checker.Check(wantEvent.Content, actualEvent.Content)
		assert.Equal(f.T(), wantContent, actualContent)
	}
}

func (f *FeedTestSuite) setupEvent(ctx context.Context, mockFollowClient *followMocks.MockFollowServiceClient, server feedv1.FeedSvcServer) error {
	// 发表文章事件：用户2发表了四篇文章，用户3发表了3篇文章
	articleEvents := []ArticleEvent{
		{
			Uid:   "2",
			Aid:   "1",
			Title: "用户2发表了文章1",
		},
		{
			Uid:   "2",
			Aid:   "2",
			Title: "用户2发表了文章2",
		},
		{
			Uid:   "2",
			Aid:   "3",
			Title: "用户2发表了文章3",
		},
		{
			Uid:   "2",
			Aid:   "4",
			Title: "用户2发表了文章4",
		},
	}
	mockFollowClient.EXPECT().GetFollowStatic(gomock.Any(), &followv1.GetFollowStaticRequest{
		Followee: 2,
	}).Return(&followv1.GetFollowStaticResponse{
		FollowStatic: &followv1.FollowStatic{
			Followers: 5,
		},
	}, nil).Times(len(articleEvents))

	for _, event := range articleEvents {
		content, _ := json.Marshal(event)
		// 保证事件顺序
		time.Sleep(1 * time.Second)
		_, err := server.CreateFeedEvent(ctx, &feedv1.CreateFeedEventRequest{
			FeedEvent: &feedv1.FeedEvent{
				Type:    service.ArticleEventName,
				Content: string(content),
			},
		})
		if err != nil {
			return err
		}
	}

	articleEvents = []ArticleEvent{
		{
			Uid:   "3",
			Aid:   "5",
			Title: "用户3发表了文章5",
		},
		{
			Uid:   "3",
			Aid:   "6",
			Title: "用户3发表了文章6",
		},
		{
			Uid:   "3",
			Aid:   "7",
			Title: "用户3发表了文章7",
		},
	}
	mockFollowClient.EXPECT().GetFollowStatic(gomock.Any(), &followv1.GetFollowStaticRequest{
		Followee: 3,
	}).Return(&followv1.GetFollowStaticResponse{
		FollowStatic: &followv1.FollowStatic{
			Followers: 2,
		},
	}, nil).Times(len(articleEvents))

	mockFollowClient.EXPECT().GetFollower(gomock.Any(), &followv1.GetFollowerRequest{
		Followee: 3,
	}).Return(&followv1.GetFollowerResponse{
		FollowRelations: []*followv1.FollowRelation{
			{
				Id:       6,
				Follower: 1,
				Followee: 3,
			},
			{
				Id:       7,
				Follower: 4,
				Followee: 3,
			},
		},
	}, nil).AnyTimes()
	for _, event := range articleEvents {
		content, _ := json.Marshal(event)
		// 保证事件顺序
		time.Sleep(1 * time.Second)
		_, err := server.CreateFeedEvent(ctx, &feedv1.CreateFeedEventRequest{
			FeedEvent: &feedv1.FeedEvent{
				Type:    service.ArticleEventName,
				Content: string(content),
			},
		})
		if err != nil {
			return err
		}
	}

	// 创建点赞事件
	likeEvents := []LikeEvent{
		{
			Liked: "1",
			Liker: "10",
			BizID: "8",
			Biz:   "article",
		},
		{
			Liked: "1",
			Liker: "11",
			BizID: "9",
			Biz:   "article",
		},
		{
			Liked: "1",
			Liker: "12",
			BizID: "10",
			Biz:   "article",
		},
	}
	for _, event := range likeEvents {
		content, _ := json.Marshal(event)
		time.Sleep(1 * time.Second)
		_, err := server.CreateFeedEvent(ctx, &feedv1.CreateFeedEventRequest{
			FeedEvent: &feedv1.FeedEvent{
				Type:    service.LikeEventName,
				Content: string(content),
			},
		})
		if err != nil {
			return err
		}
	}

	// 创建关注事件
	followEvents := []FollowEvent{
		{
			Followee: "1",
			Follower: "2",
		},
		{
			Followee: "1",
			Follower: "3",
		},
		{
			Followee: "1",
			Follower: "4",
		},
	}

	for _, event := range followEvents {
		content, _ := json.Marshal(event)
		time.Sleep(1 * time.Second)
		_, err := server.CreateFeedEvent(ctx, &feedv1.CreateFeedEventRequest{
			FeedEvent: &feedv1.FeedEvent{
				Type:    service.FollowEventName,
				Content: string(content),
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *FeedTestSuite) getFeedEvent(ctx context.Context, mockFollowClient *followMocks.MockFollowServiceClient, server feedv1.FeedSvcServer) []*feedv1.FeedEvent {
	// 先获取到关注列表
	mockFollowClient.EXPECT().GetFollowee(gomock.Any(), &followv1.GetFolloweeRequest{
		Follower: 1,
		Offset:   0,
		Limit:    200,
	}).Return(&followv1.GetFolloweeResponse{
		FollowRelations: []*followv1.FollowRelation{
			{
				Id:       1,
				Follower: 1,
				Followee: 2,
			},
			{
				Id:       6,
				Follower: 1,
				Followee: 3,
			},
			{
				Id:       8,
				Follower: 1,
				Followee: 4,
			},
			{
				Id:       9,
				Follower: 1,
				Followee: 5,
			},
			{
				Id:       10,
				Follower: 1,
				Followee: 6,
			},
		},
	}, nil).AnyTimes()
	wantArticleEvents1 := []ArticleEvent{
		{
			Uid:   "2",
			Aid:   "1",
			Title: "用户2发表了文章1",
		},
		{
			Uid:   "2",
			Aid:   "2",
			Title: "用户2发表了文章2",
		},
		{
			Uid:   "2",
			Aid:   "3",
			Title: "用户2发表了文章3",
		},
		{
			Uid:   "2",
			Aid:   "4",
			Title: "用户2发表了文章4",
		},
	}
	wantArticleEvents2 := []ArticleEvent{
		{
			Uid:   "3",
			Aid:   "5",
			Title: "用户3发表了文章5",
		},
		{
			Uid:   "3",
			Aid:   "6",
			Title: "用户3发表了文章6",
		},
		{
			Uid:   "3",
			Aid:   "7",
			Title: "用户3发表了文章7",
		},
	}
	wantLikeEvents := []LikeEvent{
		{
			Liked: "1",
			Liker: "10",
			BizID: "8",
			Biz:   "article",
		},
		{
			Liked: "1",
			Liker: "11",
			BizID: "9",
			Biz:   "article",
		},
		{
			Liked: "1",
			Liker: "12",
			BizID: "10",
			Biz:   "article",
		},
	}
	wantFollowEvents := []FollowEvent{
		{
			Followee: "1",
			Follower: "2",
		},
		{
			Followee: "1",
			Follower: "3",
		},
		{
			Followee: "1",
			Follower: "4",
		},
	}

	events := make([]*feedv1.FeedEvent, 0, 32)
	for i := len(wantFollowEvents) - 1; i >= 0; i-- {
		e := wantFollowEvents[i]
		content, _ := json.Marshal(e)
		events = append(events, &feedv1.FeedEvent{
			User: &feedv1.User{
				Id: 1,
			},
			Type:    service.FollowEventName,
			Content: string(content),
		})
	}
	for i := len(wantLikeEvents) - 1; i >= 0; i-- {
		e := wantLikeEvents[i]
		content, _ := json.Marshal(e)
		events = append(events, &feedv1.FeedEvent{
			User: &feedv1.User{
				Id: 1,
			},
			Type:    service.LikeEventName,
			Content: string(content),
		})
	}
	for i := len(wantArticleEvents2) - 1; i >= 0; i-- {
		e := wantArticleEvents2[i]
		content, _ := json.Marshal(e)
		events = append(events, &feedv1.FeedEvent{
			User: &feedv1.User{
				Id: 1,
			},
			Type:    service.ArticleEventName,
			Content: string(content),
		})
	}
	for i := len(wantArticleEvents1) - 1; i >= 0; i-- {
		e := wantArticleEvents1[i]
		content, _ := json.Marshal(e)
		events = append(events, &feedv1.FeedEvent{
			User: &feedv1.User{
				Id: 1,
			},
			Type:    service.ArticleEventName,
			Content: string(content),
		})
	}
	return events
}

func removeIdAndCtime(events []*feedv1.FeedEvent) []*feedv1.FeedEvent {
	for _, e := range events {
		e.Id = 0
		e.Ctime = 0
	}
	return events
}

func TestFeedTestSuite(t *testing.T) {
	suite.Run(t, new(FeedTestSuite))
}
