package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"webook/credit/domain"
	"webook/credit/repository"
	"webook/pkg/logger"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrSelfTransfer        = errors.New("cannot transfer to self")
)

type creditService struct {
	repo repository.CreditRepository
	l    logger.LoggerV1
}

func NewCreditService(
	repo repository.CreditRepository,
	l logger.LoggerV1,
) CreditService {
	return &creditService{
		repo: repo,
		l:    l,
	}
}

// GetBalance 获取积分余额
func (s *creditService) GetBalance(ctx context.Context, uid int64) (int64, error) {
	account, err := s.repo.GetAccount(ctx, uid)
	if errors.Is(err, repository.ErrAccountNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return account.Balance, nil
}

// GetFlows 获取积分流水
func (s *creditService) GetFlows(ctx context.Context, uid int64, offset, limit int) ([]domain.CreditFlow, error) {
	return s.repo.GetFlows(ctx, uid, offset, limit)
}

// EarnCredit 积分获取（阅读、点赞、收藏、评论等）
func (s *creditService) EarnCredit(ctx context.Context, uid int64, biz string, bizId int64) (earned int64, balance int64, msg string, err error) {
	// 1. 获取积分规则
	rule, err := s.repo.GetRule(ctx, biz)
	if err != nil {
		if errors.Is(err, repository.ErrRuleNotFound) {
			return 0, 0, "无效的业务类型", nil
		}
		return 0, 0, "", err
	}

	// 2. 检查是否已获取过（幂等）
	exists, err := s.repo.HasFlow(ctx, uid, biz, bizId)
	if err != nil {
		return 0, 0, "", err
	}
	if exists {
		// 已经获取过，返回当前余额
		balance, _ = s.GetBalance(ctx, uid)
		return 0, balance, "已获取过该积分", nil
	}

	// 3. 检查每日上限
	today := time.Now().Format("2006-01-02")
	dailyLimit, err := s.repo.GetDailyLimit(ctx, uid, biz, today)
	if err != nil && !errors.Is(err, repository.ErrDailyNotFound) {
		return 0, 0, "", err
	}

	if rule.DailyLimit > 0 && dailyLimit.TotalAmt >= rule.DailyLimit {
		// 达到每日上限
		balance, _ = s.GetBalance(ctx, uid)
		return 0, balance, fmt.Sprintf("今日%s积分已达上限", biz), nil
	}

	// 4. 计算实际可获取积分（考虑每日上限）
	earnAmt := rule.CreditAmt
	if rule.DailyLimit > 0 {
		remaining := rule.DailyLimit - dailyLimit.TotalAmt
		if earnAmt > remaining {
			earnAmt = remaining
		}
	}

	// 5. 增加积分
	newBalance, err := s.repo.AddCredit(ctx, uid, biz, bizId, earnAmt, rule.Description)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateFlow) {
			balance, _ = s.GetBalance(ctx, uid)
			return 0, balance, "已获取过该积分", nil
		}
		return 0, 0, "", err
	}

	// 6. 更新每日记录
	if err = s.repo.IncrDailyLimit(ctx, uid, biz, today, earnAmt); err != nil {
		s.l.Error("更新每日积分记录失败", logger.Error(err))
		// 非关键错误，不影响主流程
	}

	return earnAmt, newBalance, "", nil
}

// RewardCredit 积分打赏
func (s *creditService) RewardCredit(ctx context.Context, uid, targetUid int64, biz string, bizId int64, amt int64) (int64, error) {
	// 检查余额
	balance, err := s.GetBalance(ctx, uid)
	if err != nil {
		return 0, err
	}
	if balance < amt {
		return 0, errors.New("积分余额不足")
	}

	// 不能给自己打赏
	if uid == targetUid {
		return 0, errors.New("不能给自己打赏")
	}

	// 创建打赏记录
	reward := domain.CreditReward{
		Uid:       uid,
		TargetUid: targetUid,
		Biz:       biz,
		BizId:     bizId,
		Amount:    amt,
		Status:    domain.CreditRewardStatusPending,
	}
	rewardId, err := s.repo.CreateCreditReward(ctx, reward)
	if err != nil {
		return 0, err
	}

	// 执行转账（事务）
	err = s.repo.TransferCredit(ctx, uid, targetUid, amt, biz, rewardId)
	if err != nil {
		_ = s.repo.UpdateCreditRewardStatus(ctx, rewardId, domain.CreditRewardStatusFailed)
		return 0, err
	}

	// 更新打赏状态为成功
	_ = s.repo.UpdateCreditRewardStatus(ctx, rewardId, domain.CreditRewardStatusSuccess)

	return rewardId, nil
}

// GetCreditReward 获取打赏记录
func (s *creditService) GetCreditReward(ctx context.Context, rewardId, uid int64) (domain.CreditReward, error) {
	reward, err := s.repo.GetCreditReward(ctx, rewardId)
	if err != nil {
		return domain.CreditReward{}, err
	}
	// 只有打赏者或被打赏者可以查看
	if reward.Uid != uid && reward.TargetUid != uid {
		return domain.CreditReward{}, errors.New("无权查看该打赏记录")
	}
	return reward, nil
}

// GetDailyStatus 获取每日积分状态
func (s *creditService) GetDailyStatus(ctx context.Context, uid int64, biz string) ([]domain.DailyStatus, error) {
	rules, err := s.repo.GetRules(ctx)
	if err != nil {
		return nil, err
	}

	today := time.Now().Format("2006-01-02")
	var statuses []domain.DailyStatus

	for _, rule := range rules {
		// 如果指定了业务类型，只返回该类型
		if biz != "" && rule.Biz != biz {
			continue
		}

		daily, err := s.repo.GetDailyLimit(ctx, uid, rule.Biz, today)
		if err != nil && !errors.Is(err, repository.ErrDailyNotFound) {
			return nil, err
		}

		remaining := rule.DailyLimit - daily.TotalAmt
		if remaining < 0 {
			remaining = 0
		}

		statuses = append(statuses, domain.DailyStatus{
			Biz:         rule.Biz,
			EarnedCount: daily.Count,
			EarnedAmt:   daily.TotalAmt,
			DailyLimit:  rule.DailyLimit,
			Remaining:   remaining,
		})
	}

	return statuses, nil
}

// DeductCredit 扣减积分（开放API专用）
func (s *creditService) DeductCredit(ctx context.Context, uid int64, amount int64, biz, bizId, description string) error {
	// 检查余额
	balance, err := s.GetBalance(ctx, uid)
	if err != nil {
		return err
	}
	if balance < amount {
		return ErrInsufficientBalance
	}

	// 执行扣减（使用负数）
	_, err = s.repo.AddCredit(ctx, uid, biz, s.hashBizId(bizId), -amount, description)
	return err
}

// Transfer 积分转账（开放API专用，不抽成）
func (s *creditService) Transfer(ctx context.Context, fromUid, toUid, amount int64, description string) error {
	if fromUid == toUid {
		return ErrSelfTransfer
	}

	// 检查余额
	balance, err := s.GetBalance(ctx, fromUid)
	if err != nil {
		return err
	}
	if balance < amount {
		return ErrInsufficientBalance
	}

	// 执行转账（使用现有的TransferCredit方法，但需要修改为不抽成）
	return s.repo.TransferCreditFull(ctx, fromUid, toUid, amount, description)
}

func (s *creditService) hashBizId(bizId string) int64 {
	// 简单的字符串转int64
	var hash int64
	for _, c := range bizId {
		hash = hash*31 + int64(c)
	}
	return hash
}
