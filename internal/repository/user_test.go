package repository

import (
	"database/sql"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/net/context"
	"testing"
	"time"
	"webook/internal/domain"
	"webook/internal/repository/cache"
	cachemocks "webook/internal/repository/cache/mocks"
	"webook/internal/repository/dao"
	daomocks "webook/internal/repository/dao/mocks"
)

func TestCachedUserRepository_FindById(t *testing.T) {
	// 因为存储的是毫秒
	// 所以需要去掉纳秒部分
	now := time.Now()
	now = time.UnixMilli(now.UnixMilli())
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) (dao.UserDao, cache.UserCache)
		id       int64
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "缓存未命中，从DB中找到数据",
			mock: func(ctrl *gomock.Controller) (dao.UserDao, cache.UserCache) {
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), gomock.Any()).Return(domain.User{}, cache.ErrKeyNotExist)
				d := daomocks.NewMockUserDao(ctrl)
				d.EXPECT().FindById(gomock.Any(), int64(123)).Return(dao.User{
					Id:       123,
					Email:    sql.NullString{String: "123@qq.com", Valid: true},
					Password: "password",
					Phone:    sql.NullString{String: "15000000000", Valid: true},
					Ctime:    now.UnixMilli(),
					Birthday: now.UnixMilli(),
				}, nil)
				c.EXPECT().Set(gomock.Any(), domain.User{
					Id:       123,
					Email:    "123@qq.com",
					Password: "password",
					Phone:    "15000000000",
					Birthday: now,
					Ctime:    now,
				}).Return(nil)
				return d, c
			},
			id:      int64(123),
			wantErr: nil,
			wantUser: domain.User{
				Id:       123,
				Email:    "123@qq.com",
				Password: "password",
				Phone:    "15000000000",
				Ctime:    now,
				Birthday: now,
			},
		},
		{
			name: "缓存命中",
			mock: func(ctrl *gomock.Controller) (dao.UserDao, cache.UserCache) {
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), gomock.Any()).Return(domain.User{
					Id:       123,
					Email:    "123@qq.com",
					Password: "password",
					Phone:    "15000000000",
					Ctime:    now,
					Birthday: now,
				}, nil)
				d := daomocks.NewMockUserDao(ctrl)
				return d, c
			},
			id:      int64(123),
			wantErr: nil,
			wantUser: domain.User{
				Id:       123,
				Email:    "123@qq.com",
				Password: "password",
				Phone:    "15000000000",
				Ctime:    now,
				Birthday: now,
			},
		},
		{
			name: "缓存未命中，DB未命中",
			mock: func(ctrl *gomock.Controller) (dao.UserDao, cache.UserCache) {
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), gomock.Any()).Return(domain.User{}, cache.ErrKeyNotExist)
				d := daomocks.NewMockUserDao(ctrl)
				d.EXPECT().FindById(gomock.Any(), int64(123)).Return(dao.User{}, errors.New("db error"))
				return d, c
			},
			id:       int64(123),
			wantErr:  errors.New("db error"),
			wantUser: domain.User{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := NewCachedUserRepository(tc.mock(ctrl))
			user, err := repo.FindById(context.Background(), tc.id)

			assert.Equal(t, tc.wantUser, user)
			assert.Equal(t, tc.wantErr, err)
			// 测试异步，等待一秒钟，让goroutine执行
			time.Sleep(time.Second)
		})
	}
}
