package handler

import (
	"errors"
	"fmt"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
	"net/http"
	"time"
	codev1 "webook/api/proto/gen/code/v1"
	userv1 "webook/api/proto/gen/user/v1"
	"webook/internal/errs"
	ijwt "webook/internal/handler/jwt"
	"webook/pkg/ginx"
	"webook/pkg/logger"
)

const biz = "login"

type UserHandler struct {
	svc         userv1.UserServiceClient
	codeSvc     codev1.CodeServiceClient
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
	phoneExp    *regexp.Regexp
	ijwt.Handler
	cmd redis.Cmdable
	l   logger.LoggerV1
}

func NewUserHandler(svc userv1.UserServiceClient, codeSvc codev1.CodeServiceClient, cmd redis.Cmdable, handler ijwt.Handler, l logger.LoggerV1) *UserHandler {
	const (
		emailRegexPattern    = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[@$!%*?&.])[A-Za-z\d@$!%*?&.]{8,72}$`
		phoneRegexPattern    = `^1[3-9]\d{9}$`
	)
	return &UserHandler{
		svc:         svc,
		codeSvc:     codeSvc,
		emailExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		phoneExp:    regexp.MustCompile(phoneRegexPattern, regexp.None),
		Handler:     handler,
		cmd:         cmd,
		l:           l,
	}
}

func (h *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/signup", ginx.WrapBody(h.l, h.SignUp))
	ug.POST("/login", ginx.WrapBody(h.l, h.LoginJWT))
	ug.POST("/edit", ginx.WrapBody(h.l, h.Edit))
	ug.GET("/profile", ginx.WrapClaims(h.l, h.Profile))
	ug.POST("/login_sms/code/send", ginx.WrapBody(h.l, h.SendLoginSMSCode))
	ug.POST("/login_sms", ginx.WrapBody(h.l, h.LoginSMS))
	ug.POST("/logout", h.LogoutJWT)
	ug.POST("/refresh_token", h.RefreshToken)
}

func (h *UserHandler) LogoutJWT(ctx *gin.Context) {
	err := h.ClearToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "退出登录失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, ginx.Result{
		Msg: "退出登录OK",
	})
	return
}

// RefreshToken 可以同时刷新长短 Token, 用 redis 来记录是否有效，即 refresh_token 是一次性的
// 参考登录校验部分，比较 User_agent 增加安全性
func (h *UserHandler) RefreshToken(ctx *gin.Context) {
	// 只有这个接口，拿出来的才是 refresh_token，其他地方都是 access_token
	refreshToken := h.ExtractToken(ctx)
	var rc ijwt.RefreshClaims
	token, err := jwt.ParseWithClaims(refreshToken, &rc, func(token *jwt.Token) (interface{}, error) {
		return ijwt.RefreshTokenKey, nil
	})
	if err != nil || !token.Valid {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	err = h.CheckSession(ctx, rc.Ssid)
	if err != nil {
		// 要么 redis 有问题，要么已经退出登录
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// 生成新的 access_token
	err = h.SetJWTToken(ctx, rc.UserId, rc.Ssid)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		// 正常来说，msg 的部分就应该包含足够的定位信息
		zap.L().Error("设置 JWT token 异常",
			zap.Error(err),
			zap.String("Method", "UserHandler:RefreshToken"))
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{Msg: "刷新成功"})
}

type LoginSMSReq struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

func (h *UserHandler) LoginSMS(ctx *gin.Context, req LoginSMSReq) (ginx.Result, error) {
	// 验证手机号
	ok, err := h.phoneExp.MatchString(req.Phone)
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
	}
	if !ok {
		return ginx.Result{Code: 4, Msg: "手机号格式不正确"}, errors.New("手机号码格式不正确")
	}

	ver, err := h.codeSvc.Verify(ctx, &codev1.VerifyRequest{
		Biz:       biz,
		InputCode: req.Code,
		Phone:     req.Phone,
	})
	ok = ver.GetAnswer()
	// TODO 利用 grpc 来传递错误码
	//if errors.Is(err, service.ErrCodeVerifyTooManyTimes) {
	//	return ginx.Result{Code: 4, Msg: "重试次数过多，请重新发送"}, err
	//}
	if err != nil {
		//ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		//zap.L().Error("校验验证码出错", zap.Error(err),
		// 不能这么打，因为手机号码是敏感数据，不能打到日志里面
		// 打印加密后的数据
		// 脱敏， 152****1212
		//zap.String("phone", req.Phone))
		// 最多打印 Debug 级别，因为生产环境中并不开 Debug
		//zap.L().Debug("", zap.String("手机号", req.Phone))
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
	}
	if !ok {
		return ginx.Result{Code: 4, Msg: "验证码错误"}, nil
	}

	resp, err := h.svc.FindOrCreate(ctx, &userv1.FindOrCreateRequest{Phone: req.Phone})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, fmt.Errorf("登录或注册用户失败 %w", err)
	}

	user := resp.GetUser()

	if err = h.SetLoginToken(ctx, user.GetId()); err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5, Msg: "系统错误",
		})
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
	}
	return ginx.Result{Code: 2, Msg: "验证码校验通过"}, nil
}

type SendLoginSMSCodeReq struct {
	Phone string `json:"phone"`
}

func (h *UserHandler) SendLoginSMSCode(ctx *gin.Context, req SendLoginSMSCodeReq) (ginx.Result, error) {
	// 验证手机号
	ok, err := h.phoneExp.MatchString(req.Phone)
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, errors.New("匹配手机格式出错")
	}
	if !ok {
		return ginx.Result{Code: 4, Msg: "手机号格式不正确"}, nil
	}
	_, err = h.codeSvc.Send(ctx.Request.Context(), &codev1.SendRequest{
		Biz:   biz,
		Phone: req.Phone,
	})
	switch {
	case err == nil:
		return ginx.Result{Msg: "发送成功"}, nil
		// TODO 利用 grpc 来传递错误码
	//case errors.Is(err, service.ErrCodeSendTooMany):
	//	return ginx.Result{Code: 4, Msg: "发送次数过多"}, fmt.Errorf("发送太频繁：%w", err)
	default:
		return ginx.Result{Code: 5, Msg: "系统错误"}, fmt.Errorf("%w", err)
	}
}

type SignUpRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

func (h *UserHandler) SignUp(ctx *gin.Context, req SignUpRequest) (ginx.Result, error) {
	ok, err := h.emailExp.MatchString(req.Email)
	if err != nil {
		// 邮箱匹配错误
		ctx.String(http.StatusOK, "系统错误")
		return ginx.Result{Msg: "系统错误"}, errors.New("匹配邮箱错误")
	}

	if !ok {
		// 邮箱格式不正确
		ctx.String(http.StatusOK, "邮箱格式不正确")
		return ginx.Result{Msg: "邮箱格式不正确"}, nil
	}

	if req.Password != req.ConfirmPassword {
		// 两次密码不一致
		return ginx.Result{Msg: "两次密码不一致"}, nil
	}

	ok, err = h.passwordExp.MatchString(req.Password)
	if err != nil {
		return ginx.Result{Msg: "系统错误"}, errors.New("密码匹配错误")
	}

	if !ok {
		return ginx.Result{Msg: "密码格式不正确"}, nil
	}

	// 调用一下 svc 的方法
	// 直接传入 ctx 在 opentelemetry 中无效，需要传入ctx.Request.Context()
	_, err = h.svc.Signup(ctx.Request.Context(), &userv1.SignUpRequest{
		User: &userv1.User{Email: req.Email, Password: req.Password},
	})
	//if errors.Is(err, service2.ErrUserDuplicate) {
	//	// 复用
	//	span := trace.SpanFromContext(ctx.Request.Context())
	//	span.AddEvent("邮箱冲突")
	//	return ginx.Result{Msg: "邮箱已注册"}, fmt.Errorf("%s 已被注册", req.Email)
	//}
	if err != nil {
		return ginx.Result{Msg: "系统错误"}, errors.New("插入数据错误")
	}
	return ginx.Result{Msg: "注册成功"}, nil
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *UserHandler) LoginJWT(ctx *gin.Context, req LoginRequest) (ginx.Result, error) {
	resp, err := h.svc.Login(ctx, &userv1.LoginRequest{Email: req.Email, Password: req.Password})

	// TODO 利用 grpc 来传递错误码
	//if errors.Is(err, service2.ErrInvalidUserOrPassword) {
	//	return ginx.Result{
	//		Code: errs.UserInvalidOrPassword,
	//		Msg:  "用户名或密码错误",
	//	}, nil
	//}

	if err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, fmt.Errorf("登录错误 %w", err)
	}

	user := resp.GetUser()

	// 设置 token
	if err = h.SetLoginToken(ctx, user.Id); err != nil {
		return ginx.Result{Msg: "系统错误"}, fmt.Errorf("token 设置错误：%w", err)
	}

	return ginx.Result{Msg: "登录成功"}, nil
}

type EditRequest struct {
	Nickname string `json:"nickname"`
	AboutMe  string `json:"about_me"`
	Birthday string `json:"birthday"`
}

func (h *UserHandler) Edit(ctx *gin.Context, req EditRequest) (ginx.Result, error) {
	// 校验参数
	if req.Nickname == "" {
		return ginx.Result{Code: 4, Msg: "昵称不能为空"}, nil
	}
	if len(req.AboutMe) > 1024 {
		return ginx.Result{Code: 4, Msg: "关于我过长"}, nil
	}
	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		return ginx.Result{Code: 4, Msg: "时间格式不对"}, nil
	}
	uid, ok := ctx.Get("userId")
	if !ok {
		return ginx.Result{Code: 5, Msg: "系统错误"}, nil
	}
	_, err = h.svc.EditNoSensitive(ctx, &userv1.EditNoSensitiveRequest{
		User: &userv1.User{
			Id:       uid.(int64),
			NickName: req.Nickname,
			AboutMe:  req.AboutMe,
			Birthday: timestamppb.New(birthday),
		},
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, fmt.Errorf("修改个人信息出错 %d %w", uid.(int64), err)
	}
	return ginx.Result{Msg: "修改成功"}, nil
}

type Profile struct {
	Email    string
	Phone    string
	Nickname string
	AboutMe  string
	Birthday string
	Ctime    string
}

func (h *UserHandler) Profile(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	resp, err := h.svc.Profile(ctx, &userv1.ProfileRequest{Id: uc.UserId})
	if err != nil {
		return ginx.Result{Msg: "系统错误"}, nil
	}
	user := resp.GetUser()
	profile := Profile{
		Email:    user.GetEmail(),
		Phone:    user.GetPhone(),
		Nickname: user.GetNickName(),
		AboutMe:  user.GetAboutMe(),
		Birthday: user.GetBirthday().AsTime().Format(time.DateOnly),
		Ctime:    user.GetCtime().AsTime().Format(time.DateOnly),
	}
	return ginx.Result{Data: profile}, nil
}
