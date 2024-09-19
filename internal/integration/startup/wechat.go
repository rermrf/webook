package startup

import (
	"webook/internal/pkg/logger"
	"webook/internal/service/oauth2/wechat"
)

func InitWechatService(l logger.LoggerV1) wechat.Service {
	return wechat.NewService("", "", l)
}
