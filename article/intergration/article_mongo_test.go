package intergration

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"webook/article/domain"
	"webook/article/repository/dao"
	ijwt "webook/bff/handler/jwt"
	startup2 "webook/user/integration/startup"
)

type ArticleMongoTestSuite struct {
	suite.Suite
	server  *gin.Engine
	mdb     *mongo.Database
	col     *mongo.Collection
	liveCol *mongo.Collection
}

func (s *ArticleMongoTestSuite) SetupSuite() {
	//s.server = startup.InitWebServer()
	s.server = gin.Default()
	// 模拟用户登录
	s.server.Use(func(ctx *gin.Context) {
		ctx.Set("claims", &ijwt.UserClaims{
			UserId: 123,
		})
	})
	s.mdb = startup2.InitMongoDB()
	node, err := snowflake.NewNode(1)
	require.NoError(s.T(), err)
	s.col = s.mdb.Collection("articles")
	s.liveCol = s.mdb.Collection("publishedArt")
	// 使用 wire 注入
	artHdl := startup2.InitArticleHandler(dao.NewMongoArticleDao(s.mdb, node))
	artHdl.RegisterRoutes(s.server)
}

// TearDownSuite 每一次测试都会执行
func (s *ArticleMongoTestSuite) TearDownSuite() {
	_, err := s.col.DeleteMany(context.TODO(), bson.M{})
	_, err = s.liveCol.DeleteMany(context.TODO(), bson.M{})
	require.NoError(s.T(), err)
}

func (s *ArticleMongoTestSuite) TestPublish() {
	testCases := []struct {
		name string
		// 预期中的输入
		art Article

		// 集成测试准备数据
		before func(t *testing.T)
		// 集成测试验证数据
		after func(t *testing.T)

		// Http 响应码
		wantCode int
		// 我希望 Http 响应带上帖子的 id
		wantRes Result[int64]
	}{
		{
			name: "新建发布成功",
			art: Article{
				Title:   "我的标题",
				Content: "我的内容",
			},
			before: func(t *testing.T) {
				// 新建发布
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				var art dao.Article
				err := s.col.FindOne(ctx, bson.M{"title": "我的标题", "author_id": 123}).Decode(&art)
				assert.NoError(s.T(), err)
				assert.True(t, art.Id > 0)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				art.Id = 0
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, art, dao.Article{
					Title:    "我的标题",
					Content:  "我的内容",
					Status:   domain.ArticleStatusPublished.ToUint8(),
					AuthorId: 123,
				})
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Msg:  "OK",
				Data: 0,
			},
		},
		{
			// 制作库有，但是线上库没有
			name: "更新帖子发布成功",
			art: Article{
				Id:      2,
				Title:   "新的标题",
				Content: "新的内容",
			},
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				// 模拟已经存在的数据
				_, err := s.col.InsertOne(ctx, &dao.Article{
					Id:       2,
					Title:    "已经存在的标题",
					Content:  "已经存在的内容",
					Status:   domain.ArticleStatusUnPublished.ToUint8(),
					AuthorId: 123,
					Ctime:    1234,
					Utime:    5678,
				})
				assert.NoError(s.T(), err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				var art dao.Article
				err := s.col.FindOne(ctx, bson.M{"id": 2}).Decode(&art)
				assert.NoError(s.T(), err)
				assert.True(t, art.Utime > 5678)
				art.Utime = 0
				assert.Equal(t, art, dao.Article{
					Id:       2,
					Title:    "新的标题",
					Content:  "新的内容",
					Status:   domain.ArticleStatusPublished.ToUint8(),
					AuthorId: 123,
					Ctime:    1234,
				})
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Msg: "OK",
			},
		},
		{
			name: "更新帖子并且重新发表",
			art: Article{
				Id:      3,
				Title:   "新的标题",
				Content: "新的内容",
			},
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				_, err := s.col.InsertOne(ctx, &dao.Article{
					Id:       3,
					Title:    "旧的标题",
					Content:  "旧的内容",
					Status:   domain.ArticleStatusPublished.ToUint8(),
					AuthorId: 123,
					Ctime:    1234,
					Utime:    5678,
				})
				assert.NoError(s.T(), err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				var art dao.Article
				//err := s.db.Where("id = ?", 3).First(&art).Error
				err := s.col.FindOne(ctx, bson.M{"id": 3}).Decode(&art)
				assert.NoError(s.T(), err)
				assert.True(t, art.Id > 0)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				art.Id = 0
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, art, dao.Article{
					Title:    "新的标题",
					Content:  "新的内容",
					Status:   domain.ArticleStatusPublished.ToUint8(),
					AuthorId: 123,
				})
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Msg: "OK",
			},
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			// 构造请求
			tc.before(t)
			reqBody, err := json.Marshal(tc.art)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "/articles/publish", bytes.NewBuffer(reqBody))
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")
			// 执行
			resp := httptest.NewRecorder()
			s.server.ServeHTTP(resp, req)

			// 验证
			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != http.StatusOK {
				return
			}
			var webRes Result[int64]
			err = json.Unmarshal(resp.Body.Bytes(), &webRes)
			assert.NoError(t, err)
			assert.True(t, webRes.Data > 0)
			webRes.Data = 0
			assert.Equal(t, tc.wantRes, webRes)
			tc.after(t)
		})
	}
}

func (s *ArticleMongoTestSuite) TestEdit() {
	testCases := []struct {
		name string
		// 预期中的输入
		art Article

		// 集成测试准备数据
		before func(t *testing.T)
		// 集成测试验证数据
		after func(t *testing.T)

		// Http 响应码
		wantCode int
		// 我希望 Http 响应带上帖子的 id
		wantRes Result[int64]
	}{
		{
			name: "新建帖子-保存成功",
			art: Article{
				Title:   "我的标题",
				Content: "我的内容",
			},
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()

				// 验证数据库
				var art dao.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "author_id", Value: 123}}).Decode(&art)
				if err != nil {
					return
				}
				if err != nil {
					return
				}
				assert.NoError(t, err)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				assert.True(t, art.Id > 0)
				art.Ctime = 0
				art.Utime = 0
				art.Id = 0
				assert.Equal(t, dao.Article{
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusUnPublished.ToUint8(),
				}, art)
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Code: 0,
				Msg:  "OK",
				Data: 0,
			},
		},
		{
			name: "修改已有的帖子，并保存",
			art: Article{
				Id:      2,
				Title:   "新的标题",
				Content: "新的内容",
			},
			before: func(t *testing.T) {
				// 修改已有帖子，必须先在数据库中预有数据
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()

				_, err := s.col.InsertOne(ctx, &dao.Article{
					Id:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
					// 跟时间有关的测试，不是逼不得已，不要用 time.Now()
					// 因为 time.Now() 每次运行都不同，很难断言
					Ctime: 1234,
					Utime: 1234,
					// 假设这是一个已经发表的，然后你去修改，改成了没发表
					Status: domain.ArticleStatusPublished.ToUint8(),
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				// 验证数据库
				var art dao.Article
				err := s.col.FindOne(ctx, bson.M{"id": 2, "author_id": 123}).Decode(&art)
				assert.NoError(t, err)
				assert.NoError(t, err)
				// 是为了确保我更新了更新时间
				assert.True(t, art.Utime > 1234)
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       2,
					Title:    "新的标题",
					Content:  "新的内容",
					AuthorId: 123,
					Ctime:    1234,
					Status:   domain.ArticleStatusUnPublished.ToUint8(),
				}, art)
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Code: 0,
				Msg:  "OK",
				Data: 2,
			},
		},
		{
			name: "修改别人的帖子",
			art: Article{
				Id:      3,
				Title:   "新的标题",
				Content: "新的内容",
			},
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				// 修改已有帖子，必须先在数据库中预有数据
				_, err := s.col.InsertOne(ctx, &dao.Article{
					Id:      3,
					Title:   "我的标题",
					Content: "我的内容",
					// 测试模拟的用户是123，这里是789
					// 意味着你在修改别人的数据
					AuthorId: 789,
					// 跟时间有关的测试，不是逼不得已，不要用 time.Now()
					// 因为 time.Now() 每次运行都不同，很难断言
					Ctime: 1234,
					Utime: 1234,
					// 为了验证状态没有变
					Status: domain.ArticleStatusPublished.ToUint8(),
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				// 验证数据库
				var art dao.Article
				err := s.col.FindOne(ctx, bson.M{"id": 3, "author_id": 789}).Decode(&art)
				assert.NoError(t, err)
				assert.Equal(t, dao.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 789,
					Ctime:    1234,
					Utime:    1234,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}, art)
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			// 构造请求
			tc.before(t)
			reqBody, err := json.Marshal(tc.art)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "/articles/edit", bytes.NewBuffer(reqBody))
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")
			// 执行
			resp := httptest.NewRecorder()
			s.server.ServeHTTP(resp, req)

			// 验证
			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != http.StatusOK {
				return
			}
			var webRes Result[int64]
			err = json.Unmarshal(resp.Body.Bytes(), &webRes)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantRes, webRes)
			tc.after(t)
		})
	}
}

func TestMongoArticle(t *testing.T) {
	suite.Run(t, &ArticleMongoTestSuite{})
}
