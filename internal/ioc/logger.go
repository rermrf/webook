package ioc

import (
	"go.uber.org/zap"
	"webook/internal/pkg/logger"
)

func InitLogger() logger.LoggerV1 {
	l, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return logger.NewZapLogger(l)
}