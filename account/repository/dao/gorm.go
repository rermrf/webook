package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type AccountGORMDAO struct {
	db *gorm.DB
}

func NewAccountGORMDAO(db *gorm.DB) AccountDAO {
	return &AccountGORMDAO{db: db}
}

// AddActivities 一次业务里面的相关账号的余额变动
func (a *AccountGORMDAO) AddActivities(ctx context.Context, activities ...AccountActivity) error {
	// 同一个业务，应该是事务
	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 修改余额
		// 添加支付记录
		now := time.Now().UnixMilli()
		for _, act := range activities {
			// 一般在用户注册的时候就会创建好账号，但是我们没有，所以要兼容处理一下
			// 注意，系统账户是默认肯定存在的，一般是离线创建好的
			// 存在账户则更新，不存在则创建账户
			err := tx.Clauses(clause.OnConflict{
				DoUpdates: clause.Assignments(map[string]interface{}{
					"balance": gorm.Expr("balance + ?", act.Amount),
					"utime":   now,
				}),
			}).Create(&Account{
				Uid:      act.Uid,
				Account:  act.Account,
				Type:     act.AccountType,
				Balance:  act.Amount,
				Currency: act.Currency,
				Ctime:    now,
				Utime:    now,
			}).Error
			if err != nil {
				return err
			}
		}
		// 批量插入
		return tx.Model(&activities).Create(&activities).Error
	})
}
