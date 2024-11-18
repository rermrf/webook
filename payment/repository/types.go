package repository

import (
	"context"
	"time"
	"webook/payment/domain"
)

//go:generate mockgen -source=./types.go -destination=mocks/payment_mock.go --package=repomocks PaymentRepository
type PaymentRepository interface {
	AddPayMent(ctx context.Context, pmt domain.Payment) error
	UpdatePayMent(ctx context.Context, pmt domain.Payment) error
	FindExpiredPayment(ctx context.Context, offiset int, limit int, t time.Time) ([]domain.Payment, error)
	GetPayment(ctx context.Context, bizTradeNO string) (domain.Payment, error)
}
