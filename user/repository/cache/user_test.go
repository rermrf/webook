package cache

import (
	"testing"
)

func TestRedisUserCache_Get(t *testing.T) {
	//now := time.Now()
	//testCases := []struct {
	//	name     string
	//	mock     func(ctrl *gomock.Controller) redis.Cmdable
	//	id       int64
	//	wantUser domain.User
	//	wantErr  error
	//}{
	//	{
	//		name: "get成功",
	//		mock: func(ctrl *gomock.Controller) redis.Cmdable {
	//			cmd := redismocks.NewMockCmdable(ctrl)
	//			res := redis.NewCmd(context.Background())
	//			res.SetErr(nil)
	//			res.SetVal(domain.User{
	//				Id:       123,
	//				Email:    "123@qq.com",
	//				Password: "password",
	//				Phone:    "15011111111",
	//				AboutMe:  "about me",
	//				Nickname: "emoji",
	//				Birthday: now,
	//				Ctime:    now,
	//			})
	//			cmd.EXPECT().Get(gomock.Any(), int64(123)).Return(res)
	//			return cmd
	//		},
	//		id: 123,
	//		wantUser: domain.User{
	//			Id:       123,
	//			Email:    "123@qq.com",
	//			Password: "password",
	//			Phone:    "15011111111",
	//			AboutMe:  "about me",
	//			Nickname: "emoji",
	//			Birthday: now,
	//			Ctime:    now,
	//		},
	//		wantErr: nil,
	//	},
	//}
	//
	//for _, testCase := range testCases {
	//	t.Run(testCase.name, func(t *testing.T) {
	//		ctrl := gomock.NewController(t)
	//		defer ctrl.Finish()
	//
	//		uc := NewUserCache(testCase.mock(ctrl))
	//		u, err := uc.Get(context.Background(), testCase.id)
	//		assert.Equal(t, testCase.wantErr, err)
	//		assert.Equal(t, testCase.wantUser, u)
	//	})
	//}
}
