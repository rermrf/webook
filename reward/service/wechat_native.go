package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	accountv1 "webook/api/proto/gen/account/v1"
	pmtv1 "webook/api/proto/gen/payment/v1"
	"webook/pkg/logger"
	"webook/reward/domain"
	"webook/reward/repository"
)

type WechatNativeRewardService struct {
	client pmtv1.WechatPaymentServiceClient
	repo   repository.RewardRepository
	l      logger.LoggerV1
	acli   accountv1.AccountServiceClient
}

func NewWechatNativeRewardService(client pmtv1.WechatPaymentServiceClient, repo repository.RewardRepository, l logger.LoggerV1) RewardService {
	return &WechatNativeRewardService{client: client, repo: repo, l: l}
}

func (s *WechatNativeRewardService) PreReward(ctx context.Context, r domain.Reward) (domain.CodeURL, error) {
	// 缓存支付二维码，一旦发现支付成功了，就清楚二维码
	// 先查询缓存，确定是否已经创建过了打赏的预支付订单
	codeURL, err := s.repo.GetCachedCodeURL(ctx, r)
	if err == nil {
		return codeURL, nil
	}
	r.Status = domain.RewardStatusInit
	rid, err := s.repo.CreateReward(ctx, r)
	if err != nil {
		return domain.CodeURL{}, err
	}
	// rpc 调用支付服务生成微信支付的二维码
	resp, err := s.client.NativePrePay(ctx, &pmtv1.PrePayRequest{
		Amt: &pmtv1.Amount{
			Total: r.Amt,
			// 这里写死货币单位
			Currency: "CNY",
		},
		// 想办法拼出一个 biz_trade_id，业务 + 打赏id
		BizTradeNo:  fmt.Sprintf("reward-%d", rid),
		Description: fmt.Sprintf("打赏-%s", r.Target.BizName),
	})
	if err != nil {
		return domain.CodeURL{}, err
	}
	cu := domain.CodeURL{
		Rid: rid,
		URL: resp.CodeUrl,
	}
	err1 := s.repo.CachedCodeURL(ctx, cu, r)
	if err1 != nil {
		s.l.Error("缓存二维码错误", logger.Error(err1))
	}
	return cu, err
}

func (s *WechatNativeRewardService) GetReward(ctx context.Context, rid, uid int64) (domain.Reward, error) {
	// 快路径
	r, err := s.repo.GetReward(ctx, rid)
	if err != nil {
		return domain.Reward{}, err
	}
	if r.Uid != uid {
		// 说明非法
		return domain.Reward{}, errors.New("查询的打赏记录和打赏人对不上")
	}
	// 有可能，我的打赏记录，还是 Init 状态：原因：第三方支付服务卡了，payment卡了，kafka消息积压了
	// 已经是完结状态
	if r.Completed() || ctx.Value("limited") == true {
		// 已经知道你的打赏结果了
		return r, nil
	}
	// 这个时候，考虑到支付到查询结果，搞一个慢路径
	// 有可能支付了，但是 reward 本身没有收到通知
	// 直接查询 payment
	// 只能解决，支付收到了，但是 reward 没收到
	// 降级状态、限流状态、熔断状态，不要走慢路径
	resp, err := s.client.GetPayment(ctx, &pmtv1.GetPaymentRequest{
		BizTradeNo: s.bizTradeNo(r.Id),
	})
	if err != nil {
		// 直接返回从数据库查到的数据
		s.l.Error("慢路径查询支付结果失败", logger.Error(err))
		return r, nil
	}
	// 更新状态
	switch resp.Status {
	case pmtv1.PaymentStatus_PaymentStatusFailed:
		r.Status = domain.RewardStatusFailed
	case pmtv1.PaymentStatus_PaymentStatusSuccess:
		r.Status = domain.RewardStatusPayed
	case pmtv1.PaymentStatus_PaymentStatusInit:
		r.Status = domain.RewardStatusInit
	case pmtv1.PaymentStatus_PaymentStatusRefund:
		// 理论上不可能出现在这个，直接设置为失败
		r.Status = domain.RewardStatusFailed
	}
	err = s.repo.UpdateStatus(ctx, rid, r.Status)
	if err != nil {
		s.l.Error("更新本地打赏状态失败", logger.Int64("rid", r.Id), logger.Error(err))
		return r, nil
	}
	return r, nil
}

func (s *WechatNativeRewardService) UpdateReward(ctx context.Context, bizTradeNO string, status domain.RewardStatus) error {
	rid := s.toRid(bizTradeNO)
	err := s.repo.UpdateStatus(ctx, rid, status)
	if err != nil {
		return err
	}
	// 完成了支付，准备入账
	if status == domain.RewardStatusPayed {
		r, err := s.repo.GetReward(ctx, rid)
		if err != nil {
			return err
		}
		// 平台抽成
		weAmt := int64(float64(r.Amt) * 0.1)
		_, err = s.acli.Credit(ctx, &accountv1.CreditRequest{
			Biz:   "reward",
			BizId: rid,
			Items: []*accountv1.CreditItem{
				{
					AccountType: accountv1.AccountType_AccountTypeReward,
					// 虽然可能为 0，但是也要记录
					Amt:      weAmt,
					Currency: "CNY",
				},
				{
					Account:     rid,
					Uid:         r.Uid,
					AccountType: accountv1.AccountType_AccountTypeReward,
					Amt:         r.Amt - weAmt,
					Currency:    "CNY",
				},
			},
		})
		if err != nil {
			s.l.Error("入账失败了", logger.String("biz_trade_no", bizTradeNO), logger.Error(err))
			return err
		}
	}
	return nil
}

func (s *WechatNativeRewardService) bizTradeNo(id int64) string {
	return fmt.Sprintf("reward-%d", id)
}

func (s *WechatNativeRewardService) toRid(no string) int64 {
	ridStr := strings.Split(no, "-")
	val, _ := strconv.ParseInt(ridStr[1], 10, 64)
	return val
}
