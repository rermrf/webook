package cache

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"webook/bff/repository/cache/redismocks"
	redismocks2 "webook/hostory/repository/cache/redismocks"
)

func TestRedisCodeCache_Set(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) redis.Cmdable
		biz     string
		phone   string
		code    string
		wantErr error
	}{
		{
			name: "set 成功",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks2.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetErr(nil)
				res.SetVal(int64(0))
				cmd.EXPECT().Eval(gomock.Any(), luaSetCode, []string{"phone_code:login:15011111111"}, []any{"200011"}).Return(res)
				return cmd
			},
			biz:     "login",
			phone:   "15011111111",
			code:    "200011",
			wantErr: nil,
		},
		{
			name: "redis错误",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks2.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetErr(errors.New("redis error"))
				cmd.EXPECT().Eval(gomock.Any(), luaSetCode, []string{"phone_code:login:15011111111"}, []any{"200011"}).Return(res)
				return cmd
			},
			biz:     "login",
			phone:   "15011111111",
			code:    "200011",
			wantErr: errors.New("redis error"),
		},
		{
			name: "一分钟内再次发送",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks2.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetErr(nil)
				res.SetVal(int64(-1))
				cmd.EXPECT().Eval(gomock.Any(), luaSetCode, []string{"phone_code:login:15011111111"}, []any{"200011"}).Return(res)
				return cmd
			},
			biz:     "login",
			phone:   "15011111111",
			code:    "200011",
			wantErr: ErrCodeSendTooMany,
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks2.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetErr(nil)
				res.SetVal(int64(10))
				cmd.EXPECT().Eval(gomock.Any(), luaSetCode, []string{"phone_code:login:15011111111"}, []any{"200011"}).Return(res)
				return cmd
			},
			biz:     "login",
			phone:   "15011111111",
			code:    "200011",
			wantErr: errors.New("系统错误"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cc := NewCodeCache(tc.mock(ctrl))
			err := cc.Set(context.Background(), tc.biz, tc.phone, tc.code)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
