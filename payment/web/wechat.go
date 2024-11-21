package web

import (
	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"net/http"
	"webook/payment/service/wechat"
	"webook/pkg/ginx"
	"webook/pkg/logger"
)

type WechatHandler struct {
	handler   *notify.Handler
	l         logger.LoggerV1
	nativeSvc *wechat.NativePaymentService
}

func NewWechatHandler(handler *notify.Handler, l logger.LoggerV1, nativeSvc *wechat.NativePaymentService) *WechatHandler {
	return &WechatHandler{handler: handler, l: l, nativeSvc: nativeSvc}
}

func (h *WechatHandler) RegisterRoutes(server *gin.Engine) {
	server.Any("/pay/callback", ginx.Wrap(h.l, h.HandleNative))
}

func (h *WechatHandler) HandleNative(ctx *gin.Context) (ginx.Result, error) {
	// 用来接收解密后的数据
	transaction := new(payments.Transaction)
	_, err := h.handler.ParseNotifyRequest(ctx, ctx.Request, transaction)
	if err != nil {
		ctx.String(http.StatusBadRequest, "解析参数失败")
		h.l.Error("解析微信支付回调失败", logger.Error(err))
		// 这里可以进一步加监控和告警
		// 觉大概率是黑客在尝试攻击
		return ginx.Result{}, nil
	}
	// 发送到 kafka
	err = h.nativeSvc.HandleCallback(ctx, transaction)
	if err != nil {
		// 在这里触发对账
		// 说明处理回调失败了
		h.l.Error("处理微信支付回调失败", logger.Error(err), logger.String("biz_trade_no", *transaction.OutTradeNo))
		return ginx.Result{}, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}
