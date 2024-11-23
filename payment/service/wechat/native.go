package wechat

import (
	"context"
	"errors"
	"fmt"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"time"
	"webook/payment/domain"
	"webook/payment/events"
	"webook/payment/repository"
	"webook/pkg/logger"
)

var errUnknownTransactionState = errors.New("未知的微信事务状态")

type NativePaymentService struct {
	appID string
	mchID string
	// 支付通知回调 URL
	notifyURL string
	// 自己的支付记录
	repo repository.PaymentRepository

	svc      *native.NativeApiService
	producer events.Producer

	l logger.LoggerV1

	// 在微信 native 里面，分别是
	// SUCCESS：支付成功
	// REFUND：转入退款
	// NOTPAY：未支付
	// CLOSED：已关闭
	// REVOKED：已撤销（付款码支付）
	// USERPAYING：用户支付中（付款码支付）
	// PAYERROR：支付失败（其他原因，如银行卡返回失败）
	nativeCBTypeToStatus map[string]domain.PaymentStatus
}

func NewNativePaymentService(appID string, mchID string, repo repository.PaymentRepository, svc *native.NativeApiService, l logger.LoggerV1, producer events.Producer) *NativePaymentService {
	return &NativePaymentService{
		appID: appID,
		mchID: mchID,
		// 一般来说，这个都是固定的，基本不会变，除非换域名
		// 从配置文件中读取
		// 1. 测试环境 test.wechat.emoji.com
		// 2. 开发环境 dev.wechat.emoji.com
		// 3. 线上环境 wechat.emoji.com
		notifyURL: "http://wechat.ermoji.com/pay/callback",
		repo:      repo,
		svc:       svc,
		l:         l,
		producer:  producer,
		nativeCBTypeToStatus: map[string]domain.PaymentStatus{
			"SUCCESS":    domain.PaymentStatusSuccess,
			"PAYERROR":   domain.PaymentStatusFailed,
			"NOTPAY":     domain.PaymentStatusInit,
			"CLOSED":     domain.PaymentStatusFailed,
			"REVOKED":    domain.PaymentStatusFailed,
			"REFUND":     domain.PaymentStatusRefund,
			"USERPAYING": domain.PaymentStatusInit,
		},
	}
}

// Prepay 为了拿到扫码支付的二维码
func (n *NativePaymentService) Prepay(ctx context.Context, pmt domain.Payment) (string, error) {
	pmt.Status = domain.PaymentStatusInit
	// 唯一索引冲突
	// 业务方唤起了支付，但是没付，下一次再过来，应该换 BizTradeNO
	err := n.repo.AddPayMent(ctx, pmt)
	if err != nil {
		return "", err
	}
	// 业务内唯一凭证
	// sn := uuid.New().String()
	resp, _, err := n.svc.Prepay(ctx, native.PrepayRequest{
		Appid:       core.String(n.appID),
		Mchid:       core.String(n.mchID),
		Description: core.String(pmt.Description),
		// 这个地方是有讲究的
		// 选择1：业务方直接传给我，我透传，我啥也不干
		// 选择2：业务方给我它的业务标识，我自己生成一个 - 担忧出现重复
		// 注意，不管你是选择 1 还是 2，业务方都一定要传给你一个唯一标识
		// Biz + BizTradeNo 唯一，biz + biz_id
		OutTradeNo: core.String(pmt.BizTradeNO),
		// 最好把这个带上，30 分钟内有效
		TimeExpire: core.Time(time.Now().Add(time.Minute * 30)),
		NotifyUrl:  core.String(n.notifyURL),
		Amount: &native.Amount{
			Total:    core.Int64(pmt.Amt.Total),
			Currency: core.String(pmt.Amt.Currency),
		},
	})

	if err != nil {
		return "", err
	}
	return *resp.CodeUrl, nil
}

// SyncWechatInfo 兜底就是准备同步状态
func (n *NativePaymentService) SyncWechatInfo(ctx context.Context, bizTradeNO string) error {
	// 对账
	txn, _, err := n.svc.QueryOrderByOutTradeNo(ctx, native.QueryOrderByOutTradeNoRequest{
		OutTradeNo: core.String(bizTradeNO),
		Mchid:      core.String(n.mchID),
	})
	if err != nil {
		return err
	}
	return n.updateByTxn(ctx, txn)
}

func (n *NativePaymentService) FindExpiredPayment(ctx context.Context, offiset, limit int, t time.Time) ([]domain.Payment, error) {
	return n.repo.FindExpiredPayment(ctx, offiset, limit, t)
}

func (n *NativePaymentService) GetPayment(ctx context.Context, bizTradeNO string) (domain.Payment, error) {
	// 在这里，我能不能设计一个慢路径？如果要是不知道支付结果，我就去微信里面查一下
	// 或者异步查一下
	return n.repo.GetPayment(ctx, bizTradeNO)
}

func (n *NativePaymentService) HandleCallback(ctx context.Context, txn *payments.Transaction) error {
	return n.updateByTxn(ctx, txn)
}

func (n *NativePaymentService) updateByTxn(ctx context.Context, txn *payments.Transaction) error {
	// 核心就是更新数据库状态
	status, ok := n.nativeCBTypeToStatus[*txn.TradeType]
	if !ok {
		// 这个地方要告警
		return fmt.Errorf("%w，微信的状态是 %s", errUnknownTransactionState, *txn.TradeState)
	}
	// 很显然，就是更新一下我们本地数据库里面的 payment 的状态
	err := n.repo.UpdatePayMent(ctx, domain.Payment{
		// 微信过来的 transaction id
		TxnID:      *txn.TransactionId,
		BizTradeNO: *txn.OutTradeNo,
		Status:     status,
	})
	if err != nil {
		return err
	}
	// 就要通知业务方
	// 有些系统，会根据支付状态来决定要不要通知
	// 我要是发送消息失败了怎么办？
	// 站在业务的角度，你是不是至少应该成功一次
	// 这里有很多问题，核心就是部分失败问题，其次还有重复发送问题
	err = n.producer.ProducePaymentEvent(ctx, events.PaymentEvent{
		BizTradeNO: *txn.OutTradeNo,
		Status:     status.AsUint8(),
	})
	if err != nil {
		// 加监控加告警，立刻手动修复。或者自动补发
		n.l.Error("发送支付事件失败", logger.Error(err), logger.String("biz_trade_no", *txn.TransactionId))
	}
	return nil
}
