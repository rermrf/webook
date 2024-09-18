package ioc

import (
	"os"
	"webook/internal/pkg/logger"
	"webook/internal/service/oauth2/wechat"
)

func InitOAuth2WechatService(l logger.LoggerV1) wechat.Service {
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("WECHAT_APP_ID env variable is not set")
	}
	appKey, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		panic("WECHAT_APP_SECRET env variable is not set")
	}
	return wechat.NewService(appId, appKey, l)
}
