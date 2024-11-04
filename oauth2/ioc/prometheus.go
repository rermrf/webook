package ioc

import (
	"github.com/spf13/viper"
	"webook/oauth2/service/wechat"
	"webook/pkg/logger"
)

func InitPrometheus(l logger.LoggerV1) wechat.Service {
	svc := InitService(l)
	type Config struct {
		NameSpace  string `yaml:"nameSpace"`
		Subsystem  string `yaml:"subsystem"`
		InstanceId string `yaml:"instanceId"`
		Name       string `yaml:"name"`
	}
	var cfg Config
	err := viper.UnmarshalKey("prometheus", &cfg)
	if err != nil {
		panic(err)
	}
	return wechat.NewPrometheusDecorator(svc, cfg.NameSpace, cfg.Subsystem, cfg.Name, cfg.InstanceId)
}
