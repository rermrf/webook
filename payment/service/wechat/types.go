package wechat

import (
	"context"
	"webook/payment/domain"
)

type PayMentService interface {
	// Prepay 预支付，对应与微信创建订单的步骤
	Prepay(ctx context.Context, pmt domain.Payment) (string, error)
}
