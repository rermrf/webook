package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"webook/oauth2/domain"
	"webook/pkg/logger"
)

var (
	redirectURL = url.PathEscape("https://yuming.com/oauth2/wechat/callbak")
)

type Service interface {
	AuthURL(ctx context.Context, state string) (string, error)
	VerifyCode(ctx context.Context, code string, state string) (domain.WechatInfo, error)
}

type service struct {
	appId     string
	appSecret string
	//cmd       redis.Cmdable
	l logger.LoggerV1
}

func NewService(appId string, appSecret string, l logger.LoggerV1) Service {
	return &service{
		appId:     appId,
		appSecret: appSecret,
		l:         l,
	}
}

func (s *service) VerifyCode(ctx context.Context, code string, _ string) (domain.WechatInfo, error) {
	const targetPattern = "https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code"
	target := fmt.Sprintf(targetPattern, s.appId, s.appSecret, code)
	resq, err := http.Get(target)
	if err != nil {
		return domain.WechatInfo{}, err
	}

	// 只读一遍
	decoder := json.NewDecoder(resq.Body)
	var res Result
	err = decoder.Decode(&res)

	// 整个响应读出来，不推荐，因为 Unmarshal 再读一遍，合计两边
	//body, err := io.ReadAll(resq.Body)
	//err = json.Unmarshal(body, &res)

	if err != nil {
		return domain.WechatInfo{}, err
	}

	if res.ErrCode != 0 {
		return domain.WechatInfo{}, fmt.Errorf("微信返回错误响应，错误码：%d，错误信息：%s", res.ErrCode, res.ErrMsg)
	}

	//cacheState := s.cmd.Get(ctx, "my-state").String()
	//if cacheState != state {
	//	// 不相同
	//}

	s.l.Info("调用微信，拿到用户信息",
		logger.String("unionID", res.UnionId),
		logger.String("openID", res.OpenId),
	)

	return domain.WechatInfo{
		OpenId:  res.OpenId,
		UnionId: res.UnionId,
	}, nil
}

func (s *service) AuthURL(ctx context.Context, state string) (string, error) {
	const urlPattern = "https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redirect"
	//s.cmd.Set(ctx, "my-state", state, time.Minute)
	return fmt.Sprintf(urlPattern, s.appId, redirectURL, state), nil
}

type Result struct {
	ErrCode int64  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`

	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`

	OpenId  string `json:"openid"`
	Scope   string `json:"scope"`
	UnionId string `json:"unionid"`
}
