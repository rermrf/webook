package repository

import (
	"context"
	"github.com/ecodeclub/ekit/sqlx"
	"webook/internal/domain"
	"webook/internal/repository/dao"
)

var ErrWaitingSMSNotFound = dao.ErrWaitingSMSNotFound

type AsyncSMSRepository interface {
	Add(ctx context.Context, s domain.AsyncSMS) error
	PreemptWaitingSMS(ctx context.Context) (domain.AsyncSMS, error)
	ReportScheduleResult(ctx context.Context, id int64, success bool) error
}

type asyncSMSRepository struct {
	dao dao.AsyncSmsDAO
}

func NewAsyncSMSRepository() AsyncSMSRepository {
	return &asyncSMSRepository{}
}

func (a *asyncSMSRepository) Add(ctx context.Context, s domain.AsyncSMS) error {
	return a.dao.Insert(ctx, dao.AsyncSms{
		Config: sqlx.JsonColumn[dao.SmsConfig]{
			Val: dao.SmsConfig{
				Biz:     s.Biz,
				Args:    s.Args,
				Numbers: s.Numbers,
			},
			Valid: true,
		},
		RetryMax: s.RetryMax,
	})
}

func (a *asyncSMSRepository) PreemptWaitingSMS(ctx context.Context) (domain.AsyncSMS, error) {
	as, err := a.dao.GetWaitingSMS(ctx)
	if err != nil {
		return domain.AsyncSMS{}, err
	}
	return domain.AsyncSMS{
		Id:       as.Id,
		Biz:      as.Config.Val.Biz,
		Numbers:  as.Config.Val.Numbers,
		Args:     as.Config.Val.Args,
		RetryMax: as.RetryMax,
	}, nil
}

func (a *asyncSMSRepository) ReportScheduleResult(ctx context.Context, id int64, success bool) error {
	if success {
		return a.dao.MarkSuccess(ctx, id)
	}
	return a.dao.MarkFailed(ctx, id)
}
