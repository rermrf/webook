//go:build manual

package wechat

import (
	"context"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

// 手动验证，提前验证
func Test_service_manual_VerifyCode(t *testing.T) {
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("WECHAT_APP_ID env variable is not set")
	}
	appKey, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		panic("WECHAT_APP_SECRET env variable is not set")
	}
	svc := NewService(appId, appKey)
	res, err := svc.VerifyCode(context.Background(), "你的auth获取到的code", "state")
	require.NoError(t, err)
	t.Log(res)
}
