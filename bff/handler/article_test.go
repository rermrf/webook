package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
	"webook/article/domain"
	"webook/article/service"
	ijwt "webook/bff/handler/jwt"
	svcmocks "webook/bff/service/mocks"
	"webook/pkg/ginx"
	"webook/pkg/logger"
)

func TestArticleHandler_Publish(t *testing.T) {
	testcases := []struct {
		name string

		mock func(ctrl *gomock.Controller, server *gin.Engine) service.ArticleService

		reqBody string

		wantCode int
		wantRes  ginx.Result
	}{
		{
			name: "新建并发表",
			mock: func(ctrl *gomock.Controller, server *gin.Engine) service.ArticleService {
				// 模拟登录态
				server.Use(func(ctx *gin.Context) {
					ctx.Set("claims", &ijwt.UserClaims{
						UserId: 123,
					})
				})
				artSvc := svcmocks.NewMockArticleService(ctrl)
				artSvc.EXPECT().Publish(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				return artSvc
			},
			reqBody:  `{ "id": 1, "title": "我的标题", "content": "我的内容" }`,
			wantCode: http.StatusOK,
			wantRes: ginx.Result{
				Msg:  "OK",
				Data: float64(1),
			},
		},
		{
			name: "publish 失败",
			mock: func(ctrl *gomock.Controller, server *gin.Engine) service.ArticleService {
				// 模拟登录态
				server.Use(func(ctx *gin.Context) {
					ctx.Set("claims", &ijwt.UserClaims{
						UserId: 123,
					})
				})
				artSvc := svcmocks.NewMockArticleService(ctrl)
				artSvc.EXPECT().Publish(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(0), errors.New("publish 失败"))
				return artSvc
			},
			reqBody:  `{ "id": 1, "title": "我的标题", "content": "我的内容" }`,
			wantCode: http.StatusOK,
			wantRes: ginx.Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
		{
			name: "参数bind失败",
			mock: func(ctrl *gomock.Controller, server *gin.Engine) service.ArticleService {
				artSvc := svcmocks.NewMockArticleService(ctrl)
				return artSvc
			},
			reqBody:  `{ "id": 1, "title": "我的标题", "content": "我的内容",,, }`,
			wantCode: http.StatusOK,
			wantRes: ginx.Result{
				Code: 4,
				Msg:  "参数错误",
			},
		},
		{
			name: "找不到 User",
			mock: func(ctrl *gomock.Controller, server *gin.Engine) service.ArticleService {
				artSvc := svcmocks.NewMockArticleService(ctrl)
				return artSvc
			},
			reqBody:  `{ "id": 1, "title": "我的标题", "content": "我的内容" }`,
			wantCode: http.StatusOK,
			wantRes: ginx.Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := gin.Default()

			h := NewArticleHandler(tc.mock(ctrl, server), &logger.NopLogger{})
			h.RegisterRoutes(server)

			req, err := http.NewRequest(http.MethodPost, "/articles/publish", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			// 设置 JSON 格式
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()

			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != http.StatusOK {
				return
			}
			var webRes ginx.Result
			err = json.NewDecoder(resp.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantRes, webRes)
		})
	}
}
