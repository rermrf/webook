package tencent

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"os"
	"testing"
)

func TestService_Send(t *testing.T) {
	secretId, ok := os.LookupEnv("SMS_SECRET_ID")
	if !ok {
		t.Fatal("SMS_SECRET_ID not set")
	}
	secretKey, ok := os.LookupEnv("SMS_SECRET_KEY")
	if !ok {
		t.Fatal("SMS_SECRET_KEY not set")
	}

	c, err := sms.NewClient(common.NewCredential(secretId, secretKey),
		"ap-guangzhou",
		profile.NewClientProfile())
	if err != nil {
		t.Fatal(err)
	}

	s := NewService(c, "1234567890", "腾马科技")

	testCases := []struct {
		name    string
		tplId   string
		params  []string
		numbers []string
		wantErr error
	}{
		{
			name:    "发送短信",
			tplId:   "123456",
			params:  []string{"123456"},
			numbers: []string{"1234567890"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			er := s.Send(context.Background(), tc.tplId, tc.params, tc.numbers...)
			assert.Equal(t, tc.wantErr, er)
		})
	}
}
