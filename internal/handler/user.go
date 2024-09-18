package handler

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"log"
	"net/http"
	"time"
	"webook/internal/domain"
	ijwt "webook/internal/handler/jwt"
	"webook/internal/service"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
)

const biz = "login"

type UserHandler struct {
	svc         service.UserService
	codeSvc     service.CodeService
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
	phoneExp    *regexp.Regexp
	ijwt.Handler
	cmd redis.Cmdable
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService, cmd redis.Cmdable, handler ijwt.Handler) *UserHandler {
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
	}
}

func (h *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/signup", h.SignUp)
	ug.POST("/login", h.LoginJWT)
	ug.POST("/edit", h.Edit)
	ug.GET("/profile", h.Profile)
	ug.POST("/login_sms/code/send", h.SendLoginSMSCode)
	ug.POST("/login_sms", h.LoginSMS)
	ug.POST("/logout", h.LogoutJWT)
	ug.POST("/refresh_token", h.RefreshToken)
}

func (h *UserHandler) LogoutJWT(ctx *gin.Context) {
	err := h.ClearToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "退出登录失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
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
	ctx.JSON(http.StatusOK, Result{Msg: "刷新成功"})
}

func (h *UserHandler) LoginSMS(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.String(http.StatusBadRequest, "请求参数错误")
		return
	}
	// 验证手机号
	ok, err := h.phoneExp.MatchString(req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "手机号格式不正确"})
		return
	}

	ok, err = h.codeSvc.Verify(ctx, biz, req.Code, req.Phone)
	if errors.Is(err, service.ErrCodeVerifyTooManyTimes) {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "重试次数过多，请重新发送"})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		zap.L().Error("校验验证码出错", zap.Error(err),
			// 不能这么打，因为手机号码是敏感数据，不能打到日志里面
			// 打印加密后的数据
			// 脱敏， 152****1212
			zap.String("phone", req.Phone))
		// 最多打印 Debug 级别，因为生产环境中并不开 Debug
		zap.L().Debug("", zap.String("手机号", req.Phone))
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "验证码错误"})
		return
	}

	user, err := h.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5, Msg: "系统错误",
		})
		return
	}

	if err = h.SetLoginToken(ctx, user.Id); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5, Msg: "系统错误",
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Code: 2,
		Msg:  "验证码校验通过",
	})
}

func (h *UserHandler) SendLoginSMSCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}
	var req Req
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(400, Result{Code: http.StatusBadRequest, Msg: err.Error()})
		return
	}
	// 验证手机号
	ok, err := h.phoneExp.MatchString(req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{Code: 5, Msg: "系统错误"})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "手机号格式不正确"})
		return
	}
	err = h.codeSvc.Send(ctx, biz, req.Phone)
	switch {
	case err == nil:
		ctx.JSON(http.StatusOK, Result{
			Msg: "发送成功",
		})
	case errors.Is(err, service.ErrCodeSendTooMany):
		// 按照链路来说，同一个 Error 可能会记录四五遍
		// 如果一条链路都是一个人负责，那就打一遍没有问题
		// 如果按照层次不同人负责，则会出现重复打日志的情况
		zap.L().Warn("发送太频繁", zap.Error(err))
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "发送次数过多",
		})
		return
	default:
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
			Data: nil,
		})
		return
	}
}

func (h *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpRequest struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}
	var req SignUpRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, "参数格式错误")
		return
	}

	ok, err := h.emailExp.MatchString(req.Email)
	if err != nil {
		// 邮箱匹配错误
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	if !ok {
		// 邮箱格式不正确
		ctx.String(http.StatusOK, "邮箱格式不正确")
		return
	}

	if req.Password != req.ConfirmPassword {
		// 两次密码不一致
		ctx.String(http.StatusOK, "两次密码不一致")
		return
	}

	ok, err = h.passwordExp.MatchString(req.Password)
	if err != nil {
		// TODO: 记录日志
		// 密码匹配错误
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	if !ok {
		// 密码格式不正确
		ctx.String(http.StatusOK, "密码格式不正确")
		return
	}

	// 调用一下 svc 的方法
	err = h.svc.SignUp(ctx, domain.User{Email: req.Email, Password: req.Password})
	if errors.Is(err, service.ErrUserDuplicate) {
		log.Println("邮箱已注册")
		ctx.String(http.StatusOK, "邮箱已注册")
		return
	}
	if err != nil {
		log.Println("插入数据错误")
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.String(http.StatusOK, "注册成功")
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *UserHandler) LoginJWT(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, "参数格式错误")
		return
	}
	user, err := h.svc.Login(ctx, req.Email, req.Password)
	if errors.Is(err, service.ErrInvalidUserOrPassword) {
		ctx.String(http.StatusOK, "用户名或密码错误")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	// TODO: 生成 token

	if err = h.SetLoginToken(ctx, user.Id); err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.String(http.StatusOK, "登录成功")
}

//
//func (h *UserHandler) Login(ctx *gin.Context) {
//	var req LoginRequest
//	if err := ctx.ShouldBindJSON(&req); err != nil {
//		return
//	}
//	user, err := h.svc.Login(ctx, req.Email, req.Password)
//	if errors.Is(err, service.ErrInvalidUserOrPassword) {
//		ctx.String(http.StatusOK, "用户名或密码错误")
//		return
//	}
//	if err != nil {
//		ctx.String(http.StatusOK, "系统错误")
//		return
//	}
//
//	// TODO: 生成 token
//	// 设置 session
//	sess := sessions.Default(ctx)
//	sess.Set("userId", user.Id)
//	sess.Options(sessions.Options{
//		// Secure: true,
//		// HttpOnly: true,
//		MaxAge: 30,
//	})
//	err = sess.Save()
//	if err != nil {
//		return
//	}
//	ctx.String(http.StatusOK, "登录成功")
//}
//
//func (h *UserHandler) Logout(ctx *gin.Context) {
//	sess := sessions.Default(ctx)
//	sess.Options(sessions.Options{
//		// Secure: true,
//		// HttpOnly: true,
//		MaxAge: -1,
//	})
//	err := sess.Save()
//	if err != nil {
//		return
//	}
//	ctx.String(http.StatusOK, "登出成功")
//}

func (h *UserHandler) Edit(ctx *gin.Context) {
	type EditRequest struct {
		Nickname string `json:"nickname"`
		AboutMe  string `json:"about_me"`
		Birthday string `json:"birthday"`
	}
	var req EditRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "参数格式错误",
		})
		return
	}
	// 校验参数
	if req.Nickname == "" {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "昵称不能为空",
		})
		return
	}
	if len(req.AboutMe) > 1024 {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "关于我过长",
		})
		return
	}
	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "时间格式不对",
		})
		return
	}
	uid, ok := ctx.Get("userId")
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	err = h.svc.EditNoSensitive(ctx, domain.User{
		Id:       uid.(int64),
		Nickname: req.Nickname,
		AboutMe:  req.AboutMe,
		Birthday: birthday,
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "修改成功",
	})
}

func (h *UserHandler) Profile(ctx *gin.Context) {
	type Profile struct {
		Email    string
		Phone    string
		Nickname string
		AboutMe  string
		Birthday string
		Ctime    string
	}
	uid, ok := ctx.Get("userId")
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	user, err := h.svc.Profile(ctx, uid.(int64))
	if errors.Is(err, service.ErrUserNotFound) {
		ctx.String(http.StatusOK, "用户不存在")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.JSON(http.StatusOK, Profile{
		Email:    user.Email,
		Phone:    user.Phone,
		Nickname: user.Nickname,
		AboutMe:  user.AboutMe,
		Birthday: user.Birthday.Format(time.DateOnly),
		Ctime:    user.Ctime.Format(time.DateOnly),
	})
}
