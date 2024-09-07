package handler

import (
	"bytes"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"testing"
	"webook/internal/domain"
	"webook/internal/service"
	svcmocks "webook/internal/service/mocks"
)

func TestEncrypt(t *testing.T) {
	password := "hello#world123"
	encrypted, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	err = bcrypt.CompareHashAndPassword(encrypted, []byte(password))
	assert.NoError(t, err)
}

func TestNil(t *testing.T) {
	testTypeAssert(nil)

}

func testTypeAssert(c any) {
	claims := c.(*UserClaims)
	println(claims.UserId)
}

func TestUserHandler_SignUp(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) service.UserService
		reqBody  string
		wantCode int
		wantBody string
	}{
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), gomock.Any()).Return(nil)
				return userSvc
			},
			reqBody: `{ 
				"email": "123@qq.com",
				"password": "w123456.",
				"confirmPassword": "w123456."
			}`,
			wantCode: http.StatusOK,
			wantBody: "注册成功",
		},
		{
			name: "参数不对，Bind失败",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				// 在调用 service 的方法之前就被返回了
				//userSvc.EXPECT().SignUp(gomock.Any(), gomock.Any()).Return(nil)
				return userSvc
			},
			reqBody: `{ 
				"email": "123@qq.com",
				"password": "w123456
			`,
			wantCode: http.StatusBadRequest,
			wantBody: "参数格式错误",
		},
		{
			name: "邮箱格式错误",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				// 在调用 service 的方法之前就被返回了
				//userSvc.EXPECT().SignUp(gomock.Any(), gomock.Any()).Return(nil)
				return userSvc
			},
			reqBody: `{ 
				"email": "123qq.com",
				"password": "w123456.",
				"confirmPassword": "w123456."
			}`,
			wantCode: http.StatusOK,
			wantBody: "邮箱格式不正确",
		},
		{
			name: "密码不一致",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				// 在调用 service 的方法之前就被返回了
				//userSvc.EXPECT().SignUp(gomock.Any(), gomock.Any()).Return(nil)
				return userSvc
			},
			reqBody: `{ 
				"email": "123@qq.com",
				"password": "w123456",
				"confirmPassword": "w123456."
			}`,
			wantCode: http.StatusOK,
			wantBody: "两次密码不一致",
		},
		{
			name: "密码格式错误",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				// 在调用 service 的方法之前就被返回了
				//userSvc.EXPECT().SignUp(gomock.Any(), gomock.Any()).Return(nil)
				return userSvc
			},
			reqBody: `{ 
				"email": "123@qq.com",
				"password": "w123456",
				"confirmPassword": "w123456"
			}`,
			wantCode: http.StatusOK,
			wantBody: "密码格式不正确",
		},
		{
			name: "邮箱已注册",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), gomock.Any()).Return(service.ErrUserDuplicate)
				return userSvc
			},
			reqBody: `{ 
				"email": "123456@qq.com",
				"password": "w123456.",
				"confirmPassword": "w123456."
			}`,
			wantCode: http.StatusOK,
			wantBody: "邮箱已注册",
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), gomock.Any()).Return(errors.New("error"))
				return userSvc
			},
			reqBody: `{ 
				"email": "123456@qq.com",
				"password": "w123456.",
				"confirmPassword": "w123456."
			}`,
			wantCode: http.StatusOK,
			wantBody: "系统错误",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := gin.Default()
			h := NewUserHandler(tc.mock(ctrl), nil)
			h.RegisterRoutes(server)

			req, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()

			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, resp.Body.String())

		})
	}
}

func TestMock(t *testing.T) {
	// mock 使用流程
	// 1. 先初始化一个控制器
	// 2. 创建模拟的对象下
	// 3. 设计模拟调用

	// 创建一个控制 mock 的控制器
	ctrl := gomock.NewController(t)
	// 每个测试结束都要调用 Finish
	// 然后 mock 就会验证你的测试流程是否符合预期
	defer ctrl.Finish()
	usersvc := svcmocks.NewMockUserService(ctrl)
	// 开始设计一个模拟调用
	// 预期第一个是 Signup 的调用
	// 模拟的条件是 gomock.Any， gomock.Any
	// 然后返回
	usersvc.EXPECT().SignUp(gomock.Any(), gomock.Any()).Return(errors.New("模拟错误"))

	err := usersvc.SignUp(context.Background(), domain.User{
		Email: "123@qq.com",
	})
	t.Log(err)
}

func TestUserHandler_LoginJWT(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) service.UserService
		reqBody  string
		wantCode int
		wantBody string
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().Login(gomock.Any(), "123@qq.com", "w123456.").Return(domain.User{Id: 1}, nil)
				return userSvc
			},
			reqBody: `{
				"email": "123@qq.com",
				"password": "w123456."
			}`,
			wantCode: http.StatusOK,
			wantBody: "登录成功",
		},
		{
			name: "参数格式错误",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc
			},
			reqBody: `{
				"email": "123@qq.com",
				"password": "w123456.
			`,
			wantCode: http.StatusBadRequest,
			wantBody: "参数格式错误",
		},
		{
			name: "用户名或密码错误",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().Login(gomock.Any(), "123@qq.com", "w123456.").Return(domain.User{}, service.ErrInvalidUserOrPassword)
				return userSvc
			},
			reqBody: `{
				"email": "123@qq.com",
				"password": "w123456."
			}`,
			wantCode: http.StatusOK,
			wantBody: "用户名或密码错误",
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().Login(gomock.Any(), "123@qq.com", "w123456.").Return(domain.User{}, errors.New("error"))
				return userSvc
			},
			reqBody: `{
				"email": "123@qq.com",
				"password": "w123456."
			}`,
			wantCode: http.StatusOK,
			wantBody: "系统错误",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := gin.Default()
			h := NewUserHandler(tc.mock(ctrl), nil)
			h.RegisterRoutes(server)

			req, err := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, resp.Body.String())
		})
	}
}

func TestUserHandler_LoginSMS(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) (service.UserService, service.CodeService)
		reqBody  string
		wantCode int
		wantBody string
	}{
		{
			name: "验证码校验通过",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				codeSvc.EXPECT().Verify(gomock.Any(), "login", "321873", "15100000000").Return(true, nil)
				userSvc.EXPECT().FindOrCreate(gomock.Any(), "15100000000").Return(domain.User{Id: 1}, nil)
				return userSvc, codeSvc
			},
			reqBody: `{
				"phone": "15100000000",
				"code": "321873"
			}`,
			wantCode: http.StatusOK,
			wantBody: `{"code":2,"msg":"验证码校验通过"}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := gin.Default()
			h := NewUserHandler(tc.mock(ctrl))
			h.RegisterRoutes(server)

			req, err := http.NewRequest(http.MethodPost, "/users/login_sms", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, resp.Body.String())
		})
	}
}
