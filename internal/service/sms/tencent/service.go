package tencent

import (
	"context"
	"fmt"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

type Service struct {
	appId    *string
	signName *string
	client   *sms.Client
}

func NewService(client *sms.Client, appId string, signName string) *Service {
	return &Service{
		appId:    &appId,
		signName: &signName,
		client:   client,
	}
}

func (s *Service) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	req := sms.NewSendSmsRequest()
	req.SmsSdkAppId = s.appId
	req.SignName = s.signName
	req.TemplateId = &tpl
	req.PhoneNumberSet = toStringPtrSlice(numbers)
	req.TemplateParamSet = toStringPtrSlice(args)
	resp, err := s.client.SendSms(req)
	if err != nil {
		return err
	}
	for _, status := range resp.Response.SendStatusSet {
		if status.Code == nil || *(status.Code) != "Ok" {
			return fmt.Errorf("发送短信失败 %s, %s", *(status.Code), *(status.Message))
		}
	}
	return nil
}

func toStringPtrSlice(src []string) []*string {
	ptrStringSlice := make([]*string, len(src))
	for i, str := range src {
		ptrStringSlice[i] = &str
	}
	return ptrStringSlice
}
