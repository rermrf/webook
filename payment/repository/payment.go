package repository

import (
	"context"
	"time"
	"webook/payment/domain"
	"webook/payment/repository/dao"
)

type paymentRepository struct {
	dao dao.PaymentDAO
}

func newPaymentRepository(dao dao.PaymentDAO) PaymentRepository {
	return &paymentRepository{dao: dao}
}

func (p *paymentRepository) AddPayMent(ctx context.Context, pmt domain.Payment) error {
	//TODO implement me
	panic("implement me")
}

func (p *paymentRepository) UpdatePayMent(ctx context.Context, pmt domain.Payment) error {
	//TODO implement me
	panic("implement me")
}

func (p *paymentRepository) FindExpiredPayment(ctx context.Context, offiset int, limit int, t time.Time) ([]domain.Payment, error) {
	//TODO implement me
	panic("implement me")
}

func (p *paymentRepository) GetPayment(ctx context.Context, bizTradeNO string) (domain.Payment, error) {
	//TODO implement me
	panic("implement me")
}

func (p *paymentRepository) toDomain(pmt dao.Payment) domain.Payment {
	return domain.Payment{
		Amt: domain.Amount{
			Currency: pmt.Currency,
			Total:    pmt.Amt,
		},
		BizTradeNO:  pmt.BizTradeNO,
		Description: pmt.Description,
		Status:      domain.PaymentStatus(pmt.Status),
		TxnID:       pmt.TxnID.String,
	}
}

func (p *paymentRepository) toEntity(pmt domain.Payment) dao.Payment {
	return dao.Payment{
		Amt:         pmt.Amt.Total,
		Currency:    pmt.Amt.Currency,
		BizTradeNO:  pmt.BizTradeNO,
		Description: pmt.Description,
		Status:      domain.PaymentStatusInit,
	}
}
