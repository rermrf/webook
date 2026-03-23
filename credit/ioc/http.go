package ioc

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"webook/credit/api"
	"webook/credit/api/epay"
	"webook/pkg/ginx"
)

func InitHTTPServer(creditAPIHandler *api.Handler, epayHandler *epay.Handler) *ginx.Server {
	engine := gin.Default()

	// 注册积分开放API路由
	creditAPIHandler.RegisterRoutes(engine)

	// 注册易支付兼容接口路由
	epayHandler.RegisterRoutes(engine)

	// 从配置读取HTTP端口
	type Config struct {
		Port string `yaml:"port"`
	}
	var cfg = Config{
		Port: ":8102", // 默认端口
	}
	_ = viper.UnmarshalKey("http.server", &cfg)

	return &ginx.Server{
		Engine: engine,
		Addr:   cfg.Port,
	}
}
