package ioc

import (
	"webook/internal/service/sms"
	"webook/internal/service/sms/memory"
)

func InitSMSService() sms.Service {
	// 这里切换验证码发送商
	return memory.NewService()
}
