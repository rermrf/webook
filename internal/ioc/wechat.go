package ioc

import (
	"os"
	"webook/internal/service/oauth2/wechat"
)

func InitOAuth2WechatService() wechat.Service {
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("WECHAT_APP_ID env variable is not set")
	}
	appKey, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		panic("WECHAT_APP_SECRET env variable is not set")
	}
	return wechat.NewService(appId, appKey)
}
