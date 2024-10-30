package ioc

import (
	"go.uber.org/zap"
	logger2 "webook/pkg/logger"
)

func InitLogger() logger2.LoggerV1 {
	l, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return logger2.NewZapLogger(l)
}
