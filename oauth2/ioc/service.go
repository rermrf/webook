package ioc

import (
	"github.com/spf13/viper"
	"webook/oauth2/service/wechat"
	"webook/pkg/logger"
)

func InitService(l logger.LoggerV1) wechat.Service {
	type Config struct {
		AppID     string `yaml:"appId"`
		AppSecret string `yaml:"appSecret"`
	}
	var cfg Config
	err := viper.UnmarshalKey("wechatConf", &cfg)
	if err != nil {
		panic(err)
	}
	return wechat.NewService(cfg.AppID, cfg.AppSecret, l)
}
