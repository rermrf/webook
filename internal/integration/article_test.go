package integration

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
	ijwt "webook/internal/handler/jwt"
	"webook/internal/integration/startup"
	"webook/internal/repository/dao"
)

// ArticleTestSuite 测试套件
type ArticleTestSuite struct {
	suite.Suite
	server *gin.Engine
	db     *gorm.DB
}

func (s *ArticleTestSuite) SetupSuite() {
	//s.server = startup.InitWebServer()
	s.server = gin.Default()
	// 模拟用户登录
	s.server.Use(func(ctx *gin.Context) {
		ctx.Set("claims", &ijwt.UserClaims{
			UserId: 123,
		})
	})
	s.db = startup.InitDB()
	// 使用 wire 注入
	artHdl := startup.InitArticleHandler()
	artHdl.RegisterRoutes(s.server)
}

// TearDownSuite 每一次测试都会执行
func (s *ArticleTestSuite) TearDownSuite() {
	// 清空所有数据，并且自增主键恢复到 1
	s.db.Exec("TRUNCATE TABLE articles")
}

func (s *ArticleTestSuite) TestEdit() {
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
				// 验证数据库
				var art dao.Article
				err := s.db.Where("id=?", 1).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       1,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
				}, art)
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Code: 0,
				Msg:  "OK",
				Data: 1,
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

func (s *ArticleTestSuite) TestPublish() {

}

func (s *ArticleTestSuite) TestABC() {
	s.T().Log("hello，这是一个测试套件")
}

func TestArticle(t *testing.T) {
	suite.Run(t, &ArticleTestSuite{})
}

type Article struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"Data"`
}
