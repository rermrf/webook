package startup

import (
	"webook/internal/pkg/logger"
)

func InitLog() logger.LoggerV1 {
	return &logger.NopLogger{}
}
