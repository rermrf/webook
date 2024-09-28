package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"webook/internal/integration/startup"
	"webook/internal/ioc"
	"webook/internal/pkg/gin-pulgin"
)

// 集成测试
func TestUserhandler_SendLoginSMSCode(t *testing.T) {
	server := startup.InitWebServer()
	rdb := ioc.InitRedis()
	testCases := []struct {
		name     string
		before   func(t *testing.T)
		after    func(t *testing.T)
		reqBody  string
		wantCode int
		wantBody gin_pulgin.Result
	}{
		{
			name: "发送成功",
			before: func(t *testing.T) {
				// 不需要，也就是 Redis 什么数据也没有
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				// 清理数据
				val, err := rdb.GetDel(ctx, "phone_code:login:15012001200").Result()
				cancel()
				assert.NoError(t, err)
				assert.True(t, len(val) == 6)
			},
			reqBody: `{
				"phone": "15012001200"
			}`,
			wantCode: http.StatusOK,
			wantBody: gin_pulgin.Result{
				Code: 0,
				Msg:  "发送成功",
			},
		},
		{
			name: "发送太频繁",
			before: func(t *testing.T) {
				// 这个手机号已经有一个验证码了
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				// 清理数据
				_, err := rdb.Set(ctx, "phone_code:login:15012001200", "123456", time.Minute*9+time.Second*30).Result()
				cancel()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				// 清理数据
				val, err := rdb.GetDel(ctx, "phone_code:login:15012001200").Result()
				cancel()
				assert.NoError(t, err)
				// 没有被覆盖， 还是123456
				assert.Equal(t, "123456", val)
			},
			reqBody: `{
				"phone": "15012001200"
			}`,
			wantCode: http.StatusOK,
			wantBody: gin_pulgin.Result{
				Code: 4,
				Msg:  "发送次数过多",
			},
		},
		{
			name: "系统错误",
			before: func(t *testing.T) {
				// 这个手机号已经有一个验证码了，但是没有过期时间
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				// 清理数据
				_, err := rdb.Set(ctx, "phone_code:login:15012001200", "123456", 0).Result()
				cancel()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				// 清理数据
				val, err := rdb.GetDel(ctx, "phone_code:login:15012001200").Result()
				cancel()
				assert.NoError(t, err)
				// 没有被覆盖， 还是123456
				assert.Equal(t, "123456", val)
			},
			reqBody: `{
				"phone": "15012001200"
			}`,
			wantCode: http.StatusOK,
			wantBody: gin_pulgin.Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
		{
			name: "手机号码为空",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
			},
			reqBody: `{
				"phone": ""
			}`,
			wantCode: http.StatusOK,
			wantBody: gin_pulgin.Result{
				Code: 4,
				Msg:  "手机号格式不正确",
			},
		},
		{
			name: "数据格式有误",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
			},
			reqBody: `{
				"phone": `,
			wantCode: 400,
			//wantBody: handler.Result{
			//	Code: 4,
			//	Msg:  "手机号格式不正确",
			//},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)

			req, err := http.NewRequest(http.MethodPost, "/users/login_sms/code/send", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != 200 {
				return
			}
			var result gin_pulgin.Result
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)
			assert.Equal(t, tc.wantBody, result)
			tc.after(t)
		})
	}
}
