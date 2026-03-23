package repository

import (
	"context"
	"webook/credit/domain"
)

type CreditRepository interface {
	// 账户相关
	GetAccount(ctx context.Context, uid int64) (domain.CreditAccount, error)
	AddCredit(ctx context.Context, uid int64, biz string, bizId int64, changeAmt int64, desc string) (int64, error)

	// 流水相关
	GetFlows(ctx context.Context, uid int64, offset, limit int) ([]domain.CreditFlow, error)
	HasFlow(ctx context.Context, uid int64, biz string, bizId int64) (bool, error)

	// 每日限制相关
	GetDailyLimit(ctx context.Context, uid int64, biz string, date string) (domain.DailyLimit, error)
	IncrDailyLimit(ctx context.Context, uid int64, biz string, date string, amt int64) error

	// 规则相关
	GetRules(ctx context.Context) ([]domain.CreditRule, error)
	GetRule(ctx context.Context, biz string) (domain.CreditRule, error)

	// 积分打赏相关
	CreateCreditReward(ctx context.Context, reward domain.CreditReward) (int64, error)
	GetCreditReward(ctx context.Context, id int64) (domain.CreditReward, error)
	UpdateCreditRewardStatus(ctx context.Context, id int64, status domain.CreditRewardStatus) error

	// 转账
	TransferCredit(ctx context.Context, fromUid, toUid, amount int64, biz string, bizId int64) error
	// 全额转账（不抽成，开放API专用）
	TransferCreditFull(ctx context.Context, fromUid, toUid, amount int64, description string) error

	// 易支付订单相关
	CreateEpayOrder(ctx context.Context, order domain.EpayOrder) (int64, error)
	GetEpayOrder(ctx context.Context, id int64) (domain.EpayOrder, error)
	GetEpayOrderByTradeNo(ctx context.Context, tradeNo string) (domain.EpayOrder, error)
	GetEpayOrderByOutTradeNo(ctx context.Context, appId, outTradeNo string) (domain.EpayOrder, error)
	UpdateEpayOrderStatus(ctx context.Context, id int64, status domain.EpayOrderStatus) error
	UpdateEpayOrderNotify(ctx context.Context, id int64, notifyCount int, notifyTime int64) error
	ListPendingNotifyOrders(ctx context.Context, limit int) ([]domain.EpayOrder, error)
}
