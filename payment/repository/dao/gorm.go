package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
	"webook/payment/domain"
)

type PaymentGORMDAO struct {
	db *gorm.DB
}

func NewPaymentGORMDAO(db *gorm.DB) PaymentDAO {
	return &PaymentGORMDAO{db: db}
}

func (p PaymentGORMDAO) Insert(ctx context.Context, pmt Payment) error {
	//TODO implement me
	panic("implement me")
}

func (p PaymentGORMDAO) UpdateTxnIDAndStatus(ctx context.Context, bizTradeNo string, txID string, status domain.PaymentStatus) error {
	//TODO implement me
	panic("implement me")
}

func (p PaymentGORMDAO) FindExpiredPayment(ctx context.Context, offset int, limit int, t time.Time) ([]Payment, error) {
	//TODO implement me
	panic("implement me")
}

func (p PaymentGORMDAO) GetPayment(ctx context.Context, bizTradeNO string) (Payment, error) {
	//TODO implement me
	panic("implement me")
}
