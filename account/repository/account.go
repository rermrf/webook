package repository

import (
	"context"
	"time"
	"webook/account/domain"
	"webook/account/repository/dao"
)

type accountRepository struct {
	dao dao.AccountDAO
}

func NewAccountRepository(dao dao.AccountDAO) AccountRepository {
	return &accountRepository{dao: dao}
}

func (a *accountRepository) AddCredit(ctx context.Context, c domain.Credit) error {
	activities := make([]dao.AccountActivity, 0, len(c.Items))
	now := time.Now().UnixMilli()
	for _, item := range c.Items {
		activities = append(activities, dao.AccountActivity{
			Uid:         item.Uid,
			Biz:         c.Biz,
			BizId:       c.BizId,
			Account:     item.Account,
			AccountType: item.AccountType.AsUint8(),
			Amount:      item.Amt,
			Currency:    item.Currency,
			Ctime:       now,
			Utime:       now,
		})
	}
	return a.dao.AddActivities(ctx, activities...)
}
