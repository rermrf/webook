package channel

import (
	"context"
	"fmt"
	"sort"

	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

// TencentSMSProvider 腾讯云短信服务商
type TencentSMSProvider struct {
	client   *sms.Client
	appId    string
	signName string
}

func NewTencentSMSProvider(client *sms.Client, appId string, signName string) *TencentSMSProvider {
	return &TencentSMSProvider{
		client:   client,
		appId:    appId,
		signName: signName,
	}
}

func (p *TencentSMSProvider) Send(ctx context.Context, tplId string, params map[string]string, numbers ...string) error {
	req := sms.NewSendSmsRequest()
	req.SmsSdkAppId = &p.appId
	req.SignName = &p.signName
	req.TemplateId = &tplId
	req.PhoneNumberSet = toStringPtrSlice(numbers)

	// 腾讯云模板参数为有序字符串数组，按 key 排序取 value
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	paramSet := make([]*string, 0, len(params))
	for _, k := range keys {
		v := params[k]
		paramSet = append(paramSet, &v)
	}
	req.TemplateParamSet = paramSet

	resp, err := p.client.SendSms(req)
	if err != nil {
		return err
	}
	for _, status := range resp.Response.SendStatusSet {
		if status.Code == nil || *status.Code != "Ok" {
			return fmt.Errorf("发送短信失败 %s, %s", *status.Code, *status.Message)
		}
	}
	return nil
}

func toStringPtrSlice(src []string) []*string {
	res := make([]*string, len(src))
	for i := range src {
		res[i] = &src[i]
	}
	return res
}
