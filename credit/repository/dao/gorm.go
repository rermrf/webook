package dao

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrAccountNotFound = errors.New("credit account not found")
	ErrFlowNotFound    = errors.New("credit flow not found")
	ErrDailyNotFound   = errors.New("daily limit not found")
	ErrRuleNotFound    = errors.New("credit rule not found")
	ErrRewardNotFound  = errors.New("credit reward not found")
	ErrDuplicateFlow   = errors.New("duplicate credit flow")
)

type CreditGORMDAO struct {
	db *gorm.DB
}

func NewCreditGORMDAO(db *gorm.DB) CreditDAO {
	return &CreditGORMDAO{db: db}
}

// GetAccount 获取积分账户
func (d *CreditGORMDAO) GetAccount(ctx context.Context, uid int64) (CreditAccount, error) {
	var account CreditAccount
	err := d.db.WithContext(ctx).Where("uid = ?", uid).First(&account).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return CreditAccount{}, ErrAccountNotFound
	}
	return account, err
}

// CreateOrUpdateAccount 创建或更新账户余额，返回更新后的余额
func (d *CreditGORMDAO) CreateOrUpdateAccount(ctx context.Context, uid int64, changeAmt int64) (int64, error) {
	now := time.Now().UnixMilli()
	// 使用 UPSERT 模式
	err := d.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "uid"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"balance": gorm.Expr("balance + ?", changeAmt),
			"utime":   now,
		}),
	}).Create(&CreditAccount{
		Uid:     uid,
		Balance: changeAmt,
		Ctime:   now,
		Utime:   now,
	}).Error
	if err != nil {
		return 0, err
	}

	// 获取更新后的余额
	account, err := d.GetAccount(ctx, uid)
	if err != nil {
		return 0, err
	}
	return account.Balance, nil
}

// CreateFlow 创建积分流水
func (d *CreditGORMDAO) CreateFlow(ctx context.Context, flow CreditFlow) error {
	flow.Ctime = time.Now().UnixMilli()
	err := d.db.WithContext(ctx).Create(&flow).Error
	if err != nil && errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrDuplicateFlow
	}
	return err
}

// GetFlows 获取积分流水列表
func (d *CreditGORMDAO) GetFlows(ctx context.Context, uid int64, offset, limit int) ([]CreditFlow, error) {
	var flows []CreditFlow
	err := d.db.WithContext(ctx).
		Where("uid = ?", uid).
		Order("ctime DESC").
		Offset(offset).
		Limit(limit).
		Find(&flows).Error
	return flows, err
}

// HasFlow 检查是否已存在流水（用于幂等检查）
func (d *CreditGORMDAO) HasFlow(ctx context.Context, uid int64, biz string, bizId int64) (bool, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&CreditFlow{}).
		Where("uid = ? AND biz = ? AND biz_id = ?", uid, biz, bizId).
		Count(&count).Error
	return count > 0, err
}

// GetDailyLimit 获取每日限制记录
func (d *CreditGORMDAO) GetDailyLimit(ctx context.Context, uid int64, biz string, date string) (DailyLimit, error) {
	var limit DailyLimit
	err := d.db.WithContext(ctx).
		Where("uid = ? AND biz = ? AND date = ?", uid, biz, date).
		First(&limit).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return DailyLimit{}, ErrDailyNotFound
	}
	return limit, err
}

// IncrDailyLimit 增加每日积分记录
func (d *CreditGORMDAO) IncrDailyLimit(ctx context.Context, uid int64, biz string, date string, amt int64) error {
	now := time.Now().UnixMilli()
	return d.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "uid"}, {Name: "biz"}, {Name: "date"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"count":     gorm.Expr("count + 1"),
			"total_amt": gorm.Expr("total_amt + ?", amt),
			"utime":     now,
		}),
	}).Create(&DailyLimit{
		Uid:      uid,
		Biz:      biz,
		Date:     date,
		Count:    1,
		TotalAmt: amt,
		Ctime:    now,
		Utime:    now,
	}).Error
}

// GetRules 获取所有启用的规则
func (d *CreditGORMDAO) GetRules(ctx context.Context) ([]CreditRule, error) {
	var rules []CreditRule
	err := d.db.WithContext(ctx).Where("enabled = ?", true).Find(&rules).Error
	return rules, err
}

// GetRule 获取指定业务的规则
func (d *CreditGORMDAO) GetRule(ctx context.Context, biz string) (CreditRule, error) {
	var rule CreditRule
	err := d.db.WithContext(ctx).Where("biz = ? AND enabled = ?", biz, true).First(&rule).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return CreditRule{}, ErrRuleNotFound
	}
	return rule, err
}

// CreateCreditReward 创建积分打赏记录
func (d *CreditGORMDAO) CreateCreditReward(ctx context.Context, reward CreditReward) (int64, error) {
	now := time.Now().UnixMilli()
	reward.Ctime = now
	reward.Utime = now
	err := d.db.WithContext(ctx).Create(&reward).Error
	return reward.Id, err
}

// GetCreditReward 获取积分打赏记录
func (d *CreditGORMDAO) GetCreditReward(ctx context.Context, id int64) (CreditReward, error) {
	var reward CreditReward
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&reward).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return CreditReward{}, ErrRewardNotFound
	}
	return reward, err
}

// UpdateCreditRewardStatus 更新积分打赏状态
func (d *CreditGORMDAO) UpdateCreditRewardStatus(ctx context.Context, id int64, status uint8) error {
	return d.db.WithContext(ctx).Model(&CreditReward{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status": status,
			"utime":  time.Now().UnixMilli(),
		}).Error
}

// TransferCredit 积分转账（事务处理）
func (d *CreditGORMDAO) TransferCredit(ctx context.Context, fromUid, toUid, amount int64, biz string, bizId int64) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UnixMilli()

		// 1. 扣减打赏者积分
		result := tx.Model(&CreditAccount{}).
			Where("uid = ? AND balance >= ?", fromUid, amount).
			Updates(map[string]interface{}{
				"balance": gorm.Expr("balance - ?", amount),
				"utime":   now,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("insufficient balance")
		}

		// 获取扣减后的余额
		var fromAccount CreditAccount
		if err := tx.Where("uid = ?", fromUid).First(&fromAccount).Error; err != nil {
			return err
		}

		// 2. 增加被打赏者积分（扣除10%平台抽成）
		receiverAmt := amount * 90 / 100 // 90% 给作者
		err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "uid"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"balance": gorm.Expr("balance + ?", receiverAmt),
				"utime":   now,
			}),
		}).Create(&CreditAccount{
			Uid:     toUid,
			Balance: receiverAmt,
			Ctime:   now,
			Utime:   now,
		}).Error
		if err != nil {
			return err
		}

		// 获取接收者余额
		var toAccount CreditAccount
		if err := tx.Where("uid = ?", toUid).First(&toAccount).Error; err != nil {
			return err
		}

		// 3. 记录打赏者流水（支出）
		if err := tx.Create(&CreditFlow{
			Uid:         fromUid,
			Biz:         "reward_out",
			BizId:       bizId,
			ChangeAmt:   -amount,
			Balance:     fromAccount.Balance,
			Description: "积分打赏支出",
			Ctime:       now,
		}).Error; err != nil {
			return err
		}

		// 4. 记录被打赏者流水（收入）
		if err := tx.Create(&CreditFlow{
			Uid:         toUid,
			Biz:         "reward_in",
			BizId:       bizId,
			ChangeAmt:   receiverAmt,
			Balance:     toAccount.Balance,
			Description: "积分打赏收入",
			Ctime:       now,
		}).Error; err != nil {
			return err
		}

		return nil
	})
}

// TransferCreditFull 全额转账（不抽成，开放API专用）
func (d *CreditGORMDAO) TransferCreditFull(ctx context.Context, fromUid, toUid, amount int64, description string) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UnixMilli()

		// 1. 扣减转出方积分
		result := tx.Model(&CreditAccount{}).
			Where("uid = ? AND balance >= ?", fromUid, amount).
			Updates(map[string]interface{}{
				"balance": gorm.Expr("balance - ?", amount),
				"utime":   now,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("insufficient balance")
		}

		// 获取扣减后的余额
		var fromAccount CreditAccount
		if err := tx.Where("uid = ?", fromUid).First(&fromAccount).Error; err != nil {
			return err
		}

		// 2. 增加接收方积分（全额，不抽成）
		err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "uid"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"balance": gorm.Expr("balance + ?", amount),
				"utime":   now,
			}),
		}).Create(&CreditAccount{
			Uid:     toUid,
			Balance: amount,
			Ctime:   now,
			Utime:   now,
		}).Error
		if err != nil {
			return err
		}

		// 获取接收者余额
		var toAccount CreditAccount
		if err := tx.Where("uid = ?", toUid).First(&toAccount).Error; err != nil {
			return err
		}

		// 3. 记录转出方流水（支出）
		bizId := now // 用时间戳作为唯一ID
		if err := tx.Create(&CreditFlow{
			Uid:         fromUid,
			Biz:         "transfer_out",
			BizId:       bizId,
			ChangeAmt:   -amount,
			Balance:     fromAccount.Balance,
			Description: description,
			Ctime:       now,
		}).Error; err != nil {
			return err
		}

		// 4. 记录接收方流水（收入）
		if err := tx.Create(&CreditFlow{
			Uid:         toUid,
			Biz:         "transfer_in",
			BizId:       bizId,
			ChangeAmt:   amount,
			Balance:     toAccount.Balance,
			Description: description,
			Ctime:       now,
		}).Error; err != nil {
			return err
		}

		return nil
	})
}

// ========== 易支付订单相关 ==========

var ErrEpayOrderNotFound = errors.New("epay order not found")

// CreateEpayOrder 创建易支付订单
func (d *CreditGORMDAO) CreateEpayOrder(ctx context.Context, order EpayOrder) (int64, error) {
	now := time.Now().UnixMilli()
	order.Ctime = now
	order.Utime = now
	err := d.db.WithContext(ctx).Create(&order).Error
	return order.Id, err
}

// GetEpayOrder 根据ID获取易支付订单
func (d *CreditGORMDAO) GetEpayOrder(ctx context.Context, id int64) (EpayOrder, error) {
	var order EpayOrder
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&order).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return EpayOrder{}, ErrEpayOrderNotFound
	}
	return order, err
}

// GetEpayOrderByTradeNo 根据平台订单号获取易支付订单
func (d *CreditGORMDAO) GetEpayOrderByTradeNo(ctx context.Context, tradeNo string) (EpayOrder, error) {
	var order EpayOrder
	err := d.db.WithContext(ctx).Where("trade_no = ?", tradeNo).First(&order).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return EpayOrder{}, ErrEpayOrderNotFound
	}
	return order, err
}

// GetEpayOrderByOutTradeNo 根据商户订单号获取易支付订单
func (d *CreditGORMDAO) GetEpayOrderByOutTradeNo(ctx context.Context, appId, outTradeNo string) (EpayOrder, error) {
	var order EpayOrder
	err := d.db.WithContext(ctx).Where("app_id = ? AND out_trade_no = ?", appId, outTradeNo).First(&order).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return EpayOrder{}, ErrEpayOrderNotFound
	}
	return order, err
}

// UpdateEpayOrderStatus 更新易支付订单状态
func (d *CreditGORMDAO) UpdateEpayOrderStatus(ctx context.Context, id int64, status uint8) error {
	return d.db.WithContext(ctx).Model(&EpayOrder{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status": status,
			"utime":  time.Now().UnixMilli(),
		}).Error
}

// UpdateEpayOrderNotify 更新易支付订单通知信息
func (d *CreditGORMDAO) UpdateEpayOrderNotify(ctx context.Context, id int64, notifyCount int, notifyTime int64) error {
	return d.db.WithContext(ctx).Model(&EpayOrder{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"notify_count": notifyCount,
			"notify_time":  notifyTime,
			"utime":        time.Now().UnixMilli(),
		}).Error
}

// ListPendingNotifyOrders 获取待通知的订单列表
func (d *CreditGORMDAO) ListPendingNotifyOrders(ctx context.Context, limit int) ([]EpayOrder, error) {
	var orders []EpayOrder
	// 查找支付成功但未通知成功的订单，且通知次数小于5次
	err := d.db.WithContext(ctx).
		Where("status = ? AND notify_count < ?", 2, 5). // status=2 是支付成功
		Order("ctime ASC").
		Limit(limit).
		Find(&orders).Error
	return orders, err
}
