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
	now := time.Now().UnixMilli()
	pmt.Utime = now
	pmt.Ctime = now
	return p.db.WithContext(ctx).Create(&pmt).Error
}

func (p PaymentGORMDAO) UpdateTxnIDAndStatus(ctx context.Context, bizTradeNo string, txID string, status domain.PaymentStatus) error {
	return p.db.WithContext(ctx).Model(&Payment{}).
		Where("biz_trade_no = ?", bizTradeNo).
		Updates(map[string]interface{}{
			"txn_id": txID,
			"status": status.AsUint8(),
			"uime":   time.Now().UnixMilli(),
		}).Error
}

func (p PaymentGORMDAO) FindExpiredPayment(ctx context.Context, offset int, limit int, t time.Time) ([]Payment, error) {
	var res []Payment
	err := p.db.WithContext(ctx).Where("status = ? AND utime < ?", uint8(domain.PaymentStatusInit), t.UnixMilli()).
		Offset(offset).
		Limit(limit).
		Find(&res).Error
	return res, err
}

func (p PaymentGORMDAO) GetPayment(ctx context.Context, bizTradeNO string) (Payment, error) {
	var res Payment
	err := p.db.WithContext(ctx).Where("biz_trade_no = ?", bizTradeNO).First(&res).Error
	return res, err
}
