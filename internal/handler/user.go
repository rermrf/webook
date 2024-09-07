package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"
	"webook/internal/domain"
	"webook/internal/service"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const biz = "login"

type UserHandler struct {
	svc         service.UserService
	codeSvc     service.CodeService
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
	phoneExp    *regexp.Regexp
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService) *UserHandler {
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
	ug.POST("/logout", h.Logout)
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
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{Code: 4, Msg: "验证码错误"})
		return
	}

	user, err := h.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		return
	}

	if err = h.setJWTToken(ctx, user.Id); err != nil {
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
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  err.Error(),
			data: nil,
		})
		return
	default:
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
			data: nil,
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

	if err = h.setJWTToken(ctx, user.Id); err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.String(http.StatusOK, "登录成功")
}

func (h *UserHandler) setJWTToken(ctx *gin.Context, userId int64) error {
	claims := UserClaims{
		UserId:    userId,
		UserAgent: ctx.Request.UserAgent(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte("pBMH@cKP65sknQI%ijB2DzhFnvsfiyt*"))
	if err != nil {
		return err
	}
	log.Println(tokenStr)
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (h *UserHandler) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
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
	// 设置 session
	sess := sessions.Default(ctx)
	sess.Set("userId", user.Id)
	sess.Options(sessions.Options{
		// Secure: true,
		// HttpOnly: true,
		MaxAge: 30,
	})
	err = sess.Save()
	if err != nil {
		return
	}
	ctx.String(http.StatusOK, "登录成功")
}

func (h *UserHandler) Logout(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	sess.Options(sessions.Options{
		// Secure: true,
		// HttpOnly: true,
		MaxAge: -1,
	})
	err := sess.Save()
	if err != nil {
		return
	}
	ctx.String(http.StatusOK, "登出成功")
}

func (h *UserHandler) Edit(ctx *gin.Context) {
	type EditRequest struct {
		Nickname string `json:"nickname"`
		AboutMe  string `json:"about_me"`
		Birthday string `json:"birthday"`
	}
	var req EditRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
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
	}
	err = h.svc.EditNoSensitive(ctx, domain.User{
		Id:       uid.(int64),
		Nickname: req.Nickname,
		AboutMe:  req.AboutMe,
		Birthday: birthday,
	})
	if err != nil {
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
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
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

func (h *UserHandler) ProfileJWT(ctx *gin.Context) {
	c, ok := ctx.Get("claims")
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	claims, ok := c.(*UserClaims)
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	log.Println(claims.UserId)
	ctx.String(http.StatusOK, "profile: "+strconv.FormatInt(claims.UserId, 10))
}

type UserClaims struct {
	UserId    int64  `json:"userId"`
	UserAgent string `json:"userAgent"`
	jwt.RegisteredClaims
}
