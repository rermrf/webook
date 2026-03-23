package repository

import (
	"context"
	"errors"

	"webook/credit/domain"
	"webook/credit/repository/cache"
	"webook/credit/repository/dao"
)

var (
	ErrAccountNotFound   = dao.ErrAccountNotFound
	ErrDailyNotFound     = dao.ErrDailyNotFound
	ErrRuleNotFound      = dao.ErrRuleNotFound
	ErrRewardNotFound    = dao.ErrRewardNotFound
	ErrDuplicateFlow     = dao.ErrDuplicateFlow
	ErrEpayOrderNotFound = dao.ErrEpayOrderNotFound
)

type creditRepository struct {
	dao   dao.CreditDAO
	cache cache.CreditCache
}

func NewCreditRepository(dao dao.CreditDAO, cache cache.CreditCache) CreditRepository {
	return &creditRepository{dao: dao, cache: cache}
}

// GetAccount 获取积分账户
func (r *creditRepository) GetAccount(ctx context.Context, uid int64) (domain.CreditAccount, error) {
	// 先尝试从缓存获取余额
	balance, err := r.cache.GetBalance(ctx, uid)
	if err == nil {
		return domain.CreditAccount{Uid: uid, Balance: balance}, nil
	}

	// 缓存未命中，从数据库获取
	account, err := r.dao.GetAccount(ctx, uid)
	if err != nil {
		return domain.CreditAccount{}, err
	}

	// 回写缓存
	_ = r.cache.SetBalance(ctx, uid, account.Balance)

	return r.accountToDomain(account), nil
}

// AddCredit 增加积分并记录流水
func (r *creditRepository) AddCredit(ctx context.Context, uid int64, biz string, bizId int64, changeAmt int64, desc string) (int64, error) {
	// 更新账户余额
	newBalance, err := r.dao.CreateOrUpdateAccount(ctx, uid, changeAmt)
	if err != nil {
		return 0, err
	}

	// 记录流水
	flow := dao.CreditFlow{
		Uid:         uid,
		Biz:         biz,
		BizId:       bizId,
		ChangeAmt:   changeAmt,
		Balance:     newBalance,
		Description: desc,
	}
	if err = r.dao.CreateFlow(ctx, flow); err != nil {
		// 流水记录失败不影响主流程，但需要记录日志
		if !errors.Is(err, dao.ErrDuplicateFlow) {
			return newBalance, err
		}
	}

	// 删除余额缓存
	_ = r.cache.DelBalance(ctx, uid)

	return newBalance, nil
}

// GetFlows 获取积分流水
func (r *creditRepository) GetFlows(ctx context.Context, uid int64, offset, limit int) ([]domain.CreditFlow, error) {
	flows, err := r.dao.GetFlows(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}

	result := make([]domain.CreditFlow, 0, len(flows))
	for _, f := range flows {
		result = append(result, r.flowToDomain(f))
	}
	return result, nil
}

// HasFlow 检查是否已存在流水
func (r *creditRepository) HasFlow(ctx context.Context, uid int64, biz string, bizId int64) (bool, error) {
	return r.dao.HasFlow(ctx, uid, biz, bizId)
}

// GetDailyLimit 获取每日限制
func (r *creditRepository) GetDailyLimit(ctx context.Context, uid int64, biz string, date string) (domain.DailyLimit, error) {
	limit, err := r.dao.GetDailyLimit(ctx, uid, biz, date)
	if err != nil {
		return domain.DailyLimit{}, err
	}
	return r.dailyLimitToDomain(limit), nil
}

// IncrDailyLimit 增加每日积分记录
func (r *creditRepository) IncrDailyLimit(ctx context.Context, uid int64, biz string, date string, amt int64) error {
	return r.dao.IncrDailyLimit(ctx, uid, biz, date, amt)
}

// GetRules 获取所有规则
func (r *creditRepository) GetRules(ctx context.Context) ([]domain.CreditRule, error) {
	rules, err := r.dao.GetRules(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]domain.CreditRule, 0, len(rules))
	for _, rule := range rules {
		result = append(result, r.ruleToDomain(rule))
	}
	return result, nil
}

// GetRule 获取指定业务的规则
func (r *creditRepository) GetRule(ctx context.Context, biz string) (domain.CreditRule, error) {
	rule, err := r.dao.GetRule(ctx, biz)
	if err != nil {
		return domain.CreditRule{}, err
	}
	return r.ruleToDomain(rule), nil
}

// CreateCreditReward 创建积分打赏记录
func (r *creditRepository) CreateCreditReward(ctx context.Context, reward domain.CreditReward) (int64, error) {
	return r.dao.CreateCreditReward(ctx, r.creditRewardToEntity(reward))
}

// GetCreditReward 获取积分打赏记录
func (r *creditRepository) GetCreditReward(ctx context.Context, id int64) (domain.CreditReward, error) {
	reward, err := r.dao.GetCreditReward(ctx, id)
	if err != nil {
		return domain.CreditReward{}, err
	}
	return r.creditRewardToDomain(reward), nil
}

// UpdateCreditRewardStatus 更新积分打赏状态
func (r *creditRepository) UpdateCreditRewardStatus(ctx context.Context, id int64, status domain.CreditRewardStatus) error {
	return r.dao.UpdateCreditRewardStatus(ctx, id, status.AsUint8())
}

// TransferCredit 积分转账
func (r *creditRepository) TransferCredit(ctx context.Context, fromUid, toUid, amount int64, biz string, bizId int64) error {
	err := r.dao.TransferCredit(ctx, fromUid, toUid, amount, biz, bizId)
	if err != nil {
		return err
	}
	// 删除双方的余额缓存
	_ = r.cache.DelBalance(ctx, fromUid)
	_ = r.cache.DelBalance(ctx, toUid)
	return nil
}

// TransferCreditFull 全额转账（不抽成，开放API专用）
func (r *creditRepository) TransferCreditFull(ctx context.Context, fromUid, toUid, amount int64, description string) error {
	err := r.dao.TransferCreditFull(ctx, fromUid, toUid, amount, description)
	if err != nil {
		return err
	}
	// 删除双方的余额缓存
	_ = r.cache.DelBalance(ctx, fromUid)
	_ = r.cache.DelBalance(ctx, toUid)
	return nil
}

// 转换函数
func (r *creditRepository) accountToDomain(account dao.CreditAccount) domain.CreditAccount {
	return domain.CreditAccount{
		Id:      account.Id,
		Uid:     account.Uid,
		Balance: account.Balance,
		Ctime:   account.Ctime,
		Utime:   account.Utime,
	}
}

func (r *creditRepository) flowToDomain(flow dao.CreditFlow) domain.CreditFlow {
	return domain.CreditFlow{
		Id:          flow.Id,
		Uid:         flow.Uid,
		Biz:         flow.Biz,
		BizId:       flow.BizId,
		ChangeAmt:   flow.ChangeAmt,
		Balance:     flow.Balance,
		Description: flow.Description,
		Ctime:       flow.Ctime,
	}
}

func (r *creditRepository) dailyLimitToDomain(limit dao.DailyLimit) domain.DailyLimit {
	return domain.DailyLimit{
		Id:       limit.Id,
		Uid:      limit.Uid,
		Biz:      limit.Biz,
		Date:     limit.Date,
		Count:    limit.Count,
		TotalAmt: limit.TotalAmt,
		Ctime:    limit.Ctime,
		Utime:    limit.Utime,
	}
}

func (r *creditRepository) ruleToDomain(rule dao.CreditRule) domain.CreditRule {
	return domain.CreditRule{
		Id:          rule.Id,
		Biz:         rule.Biz,
		CreditAmt:   rule.CreditAmt,
		DailyLimit:  rule.DailyLimit,
		Description: rule.Description,
		Enabled:     rule.Enabled,
		Ctime:       rule.Ctime,
		Utime:       rule.Utime,
	}
}

func (r *creditRepository) creditRewardToEntity(reward domain.CreditReward) dao.CreditReward {
	return dao.CreditReward{
		Id:        reward.Id,
		Uid:       reward.Uid,
		TargetUid: reward.TargetUid,
		Biz:       reward.Biz,
		BizId:     reward.BizId,
		Amount:    reward.Amount,
		Status:    reward.Status.AsUint8(),
	}
}

func (r *creditRepository) creditRewardToDomain(reward dao.CreditReward) domain.CreditReward {
	return domain.CreditReward{
		Id:        reward.Id,
		Uid:       reward.Uid,
		TargetUid: reward.TargetUid,
		Biz:       reward.Biz,
		BizId:     reward.BizId,
		Amount:    reward.Amount,
		Status:    domain.CreditRewardStatus(reward.Status),
		Ctime:     reward.Ctime,
		Utime:     reward.Utime,
	}
}

// ========== 易支付订单相关 ==========

// CreateEpayOrder 创建易支付订单
func (r *creditRepository) CreateEpayOrder(ctx context.Context, order domain.EpayOrder) (int64, error) {
	return r.dao.CreateEpayOrder(ctx, r.epayOrderToEntity(order))
}

// GetEpayOrder 根据ID获取易支付订单
func (r *creditRepository) GetEpayOrder(ctx context.Context, id int64) (domain.EpayOrder, error) {
	order, err := r.dao.GetEpayOrder(ctx, id)
	if err != nil {
		return domain.EpayOrder{}, err
	}
	return r.epayOrderToDomain(order), nil
}

// GetEpayOrderByTradeNo 根据平台订单号获取易支付订单
func (r *creditRepository) GetEpayOrderByTradeNo(ctx context.Context, tradeNo string) (domain.EpayOrder, error) {
	order, err := r.dao.GetEpayOrderByTradeNo(ctx, tradeNo)
	if err != nil {
		return domain.EpayOrder{}, err
	}
	return r.epayOrderToDomain(order), nil
}

// GetEpayOrderByOutTradeNo 根据商户订单号获取易支付订单
func (r *creditRepository) GetEpayOrderByOutTradeNo(ctx context.Context, appId, outTradeNo string) (domain.EpayOrder, error) {
	order, err := r.dao.GetEpayOrderByOutTradeNo(ctx, appId, outTradeNo)
	if err != nil {
		return domain.EpayOrder{}, err
	}
	return r.epayOrderToDomain(order), nil
}

// UpdateEpayOrderStatus 更新易支付订单状态
func (r *creditRepository) UpdateEpayOrderStatus(ctx context.Context, id int64, status domain.EpayOrderStatus) error {
	return r.dao.UpdateEpayOrderStatus(ctx, id, uint8(status))
}

// UpdateEpayOrderNotify 更新易支付订单通知信息
func (r *creditRepository) UpdateEpayOrderNotify(ctx context.Context, id int64, notifyCount int, notifyTime int64) error {
	return r.dao.UpdateEpayOrderNotify(ctx, id, notifyCount, notifyTime)
}

// ListPendingNotifyOrders 获取待通知的订单列表
func (r *creditRepository) ListPendingNotifyOrders(ctx context.Context, limit int) ([]domain.EpayOrder, error) {
	orders, err := r.dao.ListPendingNotifyOrders(ctx, limit)
	if err != nil {
		return nil, err
	}
	result := make([]domain.EpayOrder, 0, len(orders))
	for _, order := range orders {
		result = append(result, r.epayOrderToDomain(order))
	}
	return result, nil
}

func (r *creditRepository) epayOrderToEntity(order domain.EpayOrder) dao.EpayOrder {
	return dao.EpayOrder{
		Id:          order.Id,
		TradeNo:     order.TradeNo,
		OutTradeNo:  order.OutTradeNo,
		AppId:       order.AppId,
		Uid:         order.Uid,
		Type:        order.Type,
		Name:        order.Name,
		Money:       order.Money,
		Status:      uint8(order.Status),
		NotifyURL:   order.NotifyURL,
		ReturnURL:   order.ReturnURL,
		Param:       order.Param,
		NotifyCount: order.NotifyCount,
		NotifyTime:  order.NotifyTime,
	}
}

func (r *creditRepository) epayOrderToDomain(order dao.EpayOrder) domain.EpayOrder {
	return domain.EpayOrder{
		Id:          order.Id,
		TradeNo:     order.TradeNo,
		OutTradeNo:  order.OutTradeNo,
		AppId:       order.AppId,
		Uid:         order.Uid,
		Type:        order.Type,
		Name:        order.Name,
		Money:       order.Money,
		Status:      domain.EpayOrderStatus(order.Status),
		NotifyURL:   order.NotifyURL,
		ReturnURL:   order.ReturnURL,
		Param:       order.Param,
		NotifyCount: order.NotifyCount,
		NotifyTime:  order.NotifyTime,
		Ctime:       order.Ctime,
		Utime:       order.Utime,
	}
}
