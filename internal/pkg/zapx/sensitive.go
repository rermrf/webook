package zapx

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Mycore struct {
	zapcore.Core
}

func (c Mycore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	for _, field := range fields {
		if field.Key == "phone" {
			phone := field.String
			field.String = phone[:3] + "****" + phone[7:]
		}
	}
	return c.Core.Write(entry, fields)
}

func MaskPhone(key string, value string) zap.Field {
	value = value[:3] + "****" + value[7:]
	return zap.Field{
		Key:    key,
		String: value,
	}
}
