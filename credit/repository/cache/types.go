package cache

import (
	"context"
)

type CreditCache interface {
	// 余额缓存
	GetBalance(ctx context.Context, uid int64) (int64, error)
	SetBalance(ctx context.Context, uid int64, balance int64) error
	DelBalance(ctx context.Context, uid int64) error
}
