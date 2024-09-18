package tencent

import (
	"context"
	"fmt"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"go.uber.org/zap"
	mysms "webook/internal/service/sms"
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

// Send biz 直接代表的就是 tplId
func (s *Service) Send(ctx context.Context, biz string, args []mysms.NameArg, numbers ...string) error {
	req := sms.NewSendSmsRequest()
	req.SmsSdkAppId = s.appId
	req.SignName = s.signName
	req.TemplateId = &biz
	req.PhoneNumberSet = toStringPtrSlice(numbers)

	// 解决阿里云和腾讯云参数类型不一致问题
	argPtrStringSlice := make([]*string, len(args))
	for i, arg := range args {
		argPtrStringSlice[i] = &arg.Val
	}
	req.TemplateParamSet = argPtrStringSlice
	resp, err := s.client.SendSms(req)
	zap.L().Debug("发送短信",
		zap.Any("resp", resp),
		zap.Error(err))
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
