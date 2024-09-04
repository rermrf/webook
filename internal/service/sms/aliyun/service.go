package aliyun

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"strings"
	"webook/internal/service/sms"
)

type Service struct {
	appId    *string
	signName *string
	client   *dysmsapi.Client
}

func NewService(client *dysmsapi.Client, appId string, signName string) *Service {
	return &Service{
		appId:    &appId,
		signName: &signName,
		client:   client,
	}
}

func (s Service) Send(ctx context.Context, tpl string, args []sms.NameArg, numbers ...string) error {
	req := dysmsapi.CreateSendSmsRequest()
	req.SignName = *s.signName
	req.Scheme = "https"
	req.TemplateCode = tpl
	// 传入的是json
	argsMap := make(map[string]string, len(args))
	for _, arg := range args {
		argsMap[arg.Name] = arg.Val
	}
	params, err := json.Marshal(argsMap)
	if err != nil {
		return err
	}
	req.TemplateParam = string(params)

	// 阿里云多个手机号为字符串间隔
	req.PhoneNumbers = strings.Join(numbers, ",")

	resp, err := s.client.SendSms(req)

	if err != nil {
		return err
	}
	if resp.Code != "OK" {
		return errors.New(resp.Message)
	}
	return nil
}
