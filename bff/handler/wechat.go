package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lithammer/shortuuid"
	"net/http"
	"time"
	oauth2v1 "webook/api/proto/gen/oauth2/v1"
	userv1 "webook/api/proto/gen/user/v1"
	ijwt "webook/bff/handler/jwt"
	"webook/pkg/ginx"
)

type OAuth2WechatHandler struct {
	svc oauth2v1.Oauth2ServiceClient
	ijwt.Handler
	userSvc  userv1.UserServiceClient
	stateKey []byte
	cfg      Config
}

type Config struct {
	Secure bool
	//StateKey string
}

func NewOAuth2WechatHandler(svc oauth2v1.Oauth2ServiceClient, userSvc userv1.UserServiceClient, jwtHandler ijwt.Handler) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:      svc,
		userSvc:  userSvc,
		stateKey: []byte("OdmakyjatZZcNZd&L*Y9^^iD5BM^%yBV"),
		cfg:      Config{Secure: false},
		Handler:  jwtHandler,
	}
}

func (h *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/oauth2/wechat")
	g.GET("/authurl", h.AuthURL)
	g.Any("/callback", h.Callback)
}

func (h *OAuth2WechatHandler) AuthURL(ctx *gin.Context) {
	state := shortuuid.New()
	resp, err := h.svc.AuthURL(ctx.Request.Context(), &oauth2v1.AuthURLRequest{State: state})
	url := resp.GetUrl()
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "构造扫码登录URL失败",
		})
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, StateClaims{
		State: state,
		RegisteredClaims: jwt.RegisteredClaims{
			// 过期时间，预期中一个用户完成登录的时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 10)),
		},
	})
	tokenStr, err := token.SignedString(h.stateKey)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.SetCookie("jwt-state", tokenStr, 600, "/oauth2/wechat/callback", "", h.cfg.Secure, true)
	ctx.JSON(http.StatusOK, ginx.Result{
		Data: url,
	})
}

func (h *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	code := ctx.Query("code")
	//攻击者首先弄出来一个绑定微信的临时授权码 A。
	//正常用户登录成功。
	//攻击者伪造一个页面，诱导用户点击，攻击者带着正常
	//用户的 Cookie（或者JwT token）去请求，攻击者的
	//临时授权码A去绑定。
	//结果：在系统中，攻击者可以通过微信扫码登录成
	//功，看到正常用户的数据。
	state, err, done := h.verify(ctx)
	if done {
		return
	}

	//info, err := h.svc.VerifyCode(ctx, code, state)
	info, err := h.svc.VerifyCode(ctx.Request.Context(), &oauth2v1.VerifyCodeRequest{
		Code:  code,
		State: state,
	})
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	resp, err := h.userSvc.FindOrCreateByWechat(ctx, &userv1.FindOrCreateByWechatRequest{
		Info: &userv1.WechatInfo{
			OpenId:  info.GetOpenId(),
			UnionId: info.GetUnionId(),
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	user := resp.GetUser()

	// 从userService中获取id
	if err = h.SetLoginToken(ctx, user.Id); err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	ctx.JSON(http.StatusOK, ginx.Result{
		Msg: "OK",
	})
	// 验证微信的 code
}

func (h *OAuth2WechatHandler) verify(ctx *gin.Context) (string, error, bool) {
	state := ctx.Query("state")
	ck, err := ctx.Cookie("jwt-state")
	if err != nil {
		// 做好监控，防止有人恶意攻击
		// 记录日志
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 4,
			Msg:  "登录失败",
		})
		return "", nil, true
	}

	var s StateClaims
	token, err := jwt.ParseWithClaims(ck, &s, func(token *jwt.Token) (interface{}, error) {
		return h.stateKey, nil
	})
	if err != nil || !token.Valid {
		// 做好监控，防止有人恶意攻击
		// 记录日志
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 4,
			Msg:  "登录失败",
		})
		return "", nil, true
	}

	// 校验 state
	if state != s.State {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 4,
			Msg:  "登录失败",
		})
		// 做好监控，防止有人恶意攻击
		// 记录日志
		return "", nil, true
	}
	return state, err, false
}

type StateClaims struct {
	State string
	jwt.RegisteredClaims
}

//type OAuth2Handler struct {
//	wechatService
//	dingdingService
//	feishuService
//}
//
//func (h *OAuth2Handler) RegisterRoutes(server *gin.Engine) {
//	g := server.Group("/oauth2")
//	g.GET("/:platform/authurl", h.AuthURL)
//	g.Any("/:platform/callback", h.Callback)
//}
//
//func (h *OAuth2Handler) AuthURL(ctx *gin.Context) {
//	platform := ctx.Param("platform")
//	switch platform {
//	case "wechat":
//		h.wechatService.AuthURL()
//	case "dingding":
//		h.dingdingService.AuthURL()
//	case "feishu":
//		h.feishuService.AuthURL()
//	default:
//		return
//	}
//}
//
//func (h *OAuth2Handler) Callback(ctx *gin.Context) {
//
//}
