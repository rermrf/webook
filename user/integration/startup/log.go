package startup

import (
	logger2 "webook/pkg/logger"
)

func InitLog() logger2.LoggerV1 {
	return &logger2.NopLogger{}
}
