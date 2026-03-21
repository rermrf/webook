package channel

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
)

// AliyunSMSProvider 阿里云短信服务商
type AliyunSMSProvider struct {
	client   *dysmsapi.Client
	signName string
}

func NewAliyunSMSProvider(client *dysmsapi.Client, signName string) *AliyunSMSProvider {
	return &AliyunSMSProvider{
		client:   client,
		signName: signName,
	}
}

func (p *AliyunSMSProvider) Send(ctx context.Context, tplId string, params map[string]string, numbers ...string) error {
	req := dysmsapi.CreateSendSmsRequest()
	req.SignName = p.signName
	req.Scheme = "https"
	req.TemplateCode = tplId
	// 阿里云模板参数为 JSON 字符串
	paramJSON, err := json.Marshal(params)
	if err != nil {
		return err
	}
	req.TemplateParam = string(paramJSON)
	// 阿里云多个手机号用逗号分隔
	req.PhoneNumbers = strings.Join(numbers, ",")

	resp, err := p.client.SendSms(req)
	if err != nil {
		return err
	}
	if resp.Code != "OK" {
		return errors.New(resp.Message)
	}
	return nil
}
