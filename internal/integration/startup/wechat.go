package startup

import (
	"webook/oauth2/service/wechat"
	"webook/pkg/logger"
)

func InitWechatService(l logger.LoggerV1) wechat.Service {
	return wechat.NewService("", "", l)
}
