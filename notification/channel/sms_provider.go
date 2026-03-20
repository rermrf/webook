package channel

import (
	"context"
)

// SMSProvider 短信服务商接口
type SMSProvider interface {
	Send(ctx context.Context, tplId string, params map[string]string, numbers ...string) error
}

// AliyunSMSProvider 阿里云短信服务商（占位实现）
// 实际阿里云 SDK 集成将参照 sms/service/aliyun/ 的模式。
type AliyunSMSProvider struct{}

func NewAliyunSMSProvider() SMSProvider {
	return &AliyunSMSProvider{}
}

func (p *AliyunSMSProvider) Send(ctx context.Context, tplId string, params map[string]string, numbers ...string) error {
	// TODO: integrate aliyun SDK
	return nil
}
