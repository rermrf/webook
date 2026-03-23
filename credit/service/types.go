package service

import (
	"context"
	"webook/credit/domain"
)

//go:generate mockgen -source=./types.go -destination=mocks/credit_mock.go -package=svcmocks CreditService
type CreditService interface {
	// 账户相关
	GetBalance(ctx context.Context, uid int64) (int64, error)
	GetFlows(ctx context.Context, uid int64, offset, limit int) ([]domain.CreditFlow, error)

	// 积分获取
	EarnCredit(ctx context.Context, uid int64, biz string, bizId int64) (earned int64, balance int64, msg string, err error)

	// 积分打赏
	RewardCredit(ctx context.Context, uid, targetUid int64, biz string, bizId int64, amt int64) (int64, error)
	GetCreditReward(ctx context.Context, rewardId, uid int64) (domain.CreditReward, error)

	// 每日状态
	GetDailyStatus(ctx context.Context, uid int64, biz string) ([]domain.DailyStatus, error)

	// 开放API专用方法
	DeductCredit(ctx context.Context, uid int64, amount int64, biz, bizId, description string) error
	Transfer(ctx context.Context, fromUid, toUid, amount int64, description string) error
}
