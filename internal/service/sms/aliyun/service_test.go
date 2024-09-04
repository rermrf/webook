package aliyun

import (
	"context"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"github.com/stretchr/testify/assert"
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
	c := credentials.NewAccessKeyCredential(secretId, secretKey)
	client, err := dysmsapi.NewClientWithOptions("cn-nanchang", sdk.NewConfig(), c)
	if err != nil {
		t.Fatal(err)
	}

	s := NewService(client, "", "腾马科技")
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
