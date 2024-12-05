package ioc

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"webook/pkg/logger"
)

func InitLogger() logger.LoggerV1 {
	lumberLogger := &lumberjack.Logger{
		Filename:   "E:\\app\\misc\\新建文件夹\\comment.log", // 指定日志文件路径
		MaxSize:    50,                                  // 每个日志文件的最大大小
		MaxBackups: 3,                                   // 保留旧日志的最大个数
		MaxAge:     28,                                  // 保留久日志文件的最大天数
		//Compress:   true,                                // 是否压缩旧的日志文件，测试环境下不用开
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(lumberLogger),
		zapcore.DebugLevel, // 设置日志级别
	)
	l := zap.New(core, zap.AddCaller())

	return logger.NewZapLogger(l)
}
