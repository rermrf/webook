package channel

import (
	"context"
)

// SMSProvider 短信服务商接口
type SMSProvider interface {
	Send(ctx context.Context, tplId string, params map[string]string, numbers ...string) error
}
