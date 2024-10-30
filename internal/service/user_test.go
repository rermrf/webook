package service

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"log"
	"testing"
	"webook/internal/domain"
	"webook/internal/repository"
	repomocks "webook/internal/repository/mocks"
	"webook/pkg/logger"
)

func TestUserServiceImpl_SignUp(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) repository.UserRepository
		context context.Context
		user    domain.User
		wantErr error
	}{
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
				return repo
			},
			context: context.Background(),
			user:    domain.User{},
			wantErr: nil,
		},
		{
			name: "注册失败",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("error"))
				return repo
			},
			context: context.Background(),
			user:    domain.User{},
			wantErr: errors.New("error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewUserService(tc.mock(ctrl), logger.NopLogger{})
			err := svc.SignUp(tc.context, tc.user)

			assert.Equal(t, tc.wantErr, err)

		})
	}
}

func TestUserServiceImpl_Login(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) repository.UserRepository
		context  context.Context
		email    string
		password string
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), gomock.Any()).Return(domain.User{Email: "123@qq.com", Password: "$2a$10$IXcSBMGRCBpB7SY86hGZIugAghHJorWsmvwwWs5dunAPjkULCRVW."}, nil)
				return repo
			},
			context:  context.Background(),
			email:    "123@qq.com",
			password: "w12345678.",
			wantUser: domain.User{Email: "123@qq.com", Password: "$2a$10$IXcSBMGRCBpB7SY86hGZIugAghHJorWsmvwwWs5dunAPjkULCRVW."},
			wantErr:  nil,
		},
		{
			name: "用户不存在",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), gomock.Any()).Return(domain.User{}, repository.ErrUserNotFound)
				return repo
			},
			context:  context.Background(),
			email:    "123@qq.com",
			password: "w12345678.",
			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
		{
			name: "账号或密码不匹配",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), gomock.Any()).Return(domain.User{Email: "123@qq.com", Password: "w123556666."}, nil)
				return repo
			},
			context:  context.Background(),
			email:    "123@qq.com",
			password: "w12345678.",
			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
		{
			name: "数据库查询错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), gomock.Any()).Return(domain.User{}, errors.New("error"))
				return repo
			},
			context:  context.Background(),
			email:    "123@qq.com",
			password: "w12345678.",
			wantUser: domain.User{},
			wantErr:  errors.New("error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewUserService(tc.mock(ctrl), logger.NopLogger{})
			gotUser, err := svc.Login(tc.context, tc.email, tc.password)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, gotUser)
		})
	}

}

func TestEntrypted(t *testing.T) {
	res, err := bcrypt.GenerateFromPassword([]byte("w12345678."), bcrypt.DefaultCost)
	assert.NoError(t, err)
	log.Println(string(res))
}

func TestUserServiceImpl_Profile(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) repository.UserRepository
		id       int64
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "查询成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindById(gomock.Any(), gomock.Any()).Return(domain.User{Id: 1, Email: "123@qq.com"}, nil)
				return repo
			},
			id:       int64(1),
			wantUser: domain.User{Id: 1, Email: "123@qq.com"},
			wantErr:  nil,
		},
		{
			name: "查询失败",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindById(gomock.Any(), gomock.Any()).Return(domain.User{}, errors.New("error"))
				return repo
			},
			id:       int64(1),
			wantUser: domain.User{},
			wantErr:  errors.New("error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewUserService(tc.mock(ctrl), logger.NopLogger{})
			user, err := svc.Profile(context.Background(), tc.id)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}

func TestUserServiceImpl_EditNoSensitive(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) repository.UserRepository
		user    domain.User
		wantErr error
	}{
		{
			name: "修改成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().UpdateNoSensitiveById(gomock.Any(), gomock.Any()).Return(nil)
				return repo
			},
			user:    domain.User{Id: 1, Email: "123@qq.com", Nickname: "emoji", AboutMe: "about me"},
			wantErr: nil,
		},
		{
			name: "修改失败",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().UpdateNoSensitiveById(gomock.Any(), gomock.Any()).Return(errors.New("error"))
				return repo
			},
			user:    domain.User{Id: 1, Email: "123@qq.com", Nickname: "emoji", AboutMe: "about me"},
			wantErr: errors.New("error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewUserService(tc.mock(ctrl), logger.NopLogger{})
			err := svc.EditNoSensitive(context.Background(), tc.user)

			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func TestUserServiceImpl_FindOrCreate(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) repository.UserRepository
		phone    string
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "用户查找成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByPhone(gomock.Any(), gomock.Any()).Return(domain.User{Id: 1, Email: "123@qq.com", Phone: "15010000000"}, nil)
				return repo
			},
			wantUser: domain.User{Id: 1, Email: "123@qq.com", Phone: "15010000000"},
			wantErr:  nil,
		},
		{
			name: "用户查找失败但不是用户不存在",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByPhone(gomock.Any(), gomock.Any()).Return(domain.User{}, errors.New("error"))
				return repo
			},
			wantUser: domain.User{},
			wantErr:  errors.New("error"),
		},
		{
			name: "用户不存在，用户注册成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByPhone(gomock.Any(), gomock.Any()).Return(domain.User{}, repository.ErrUserNotFound)
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
				repo.EXPECT().FindByPhone(gomock.Any(), gomock.Any()).Return(domain.User{Id: 1, Email: "123@qq.com", Phone: "15010000000"}, nil)
				return repo
			},
			wantUser: domain.User{
				Id:    1,
				Email: "123@qq.com",
				Phone: "15010000000",
			},
			wantErr: nil,
		},
		{
			name: "用户不存在，但用户注册出错",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByPhone(gomock.Any(), gomock.Any()).Return(domain.User{}, repository.ErrUserNotFound)
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("error"))
				return repo
			},
			wantUser: domain.User{},
			wantErr:  errors.New("error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewUserService(tc.mock(ctrl), logger.NopLogger{})
			gotUser, err := svc.FindOrCreate(context.Background(), tc.phone)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, gotUser)
		})
	}
}
