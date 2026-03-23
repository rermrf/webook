package epay

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	openapiv1 "webook/api/proto/gen/openapi/v1"
	"webook/credit/domain"
	"webook/credit/repository"
	"webook/pkg/logger"
)

// Handler 易支付接口处理器
type Handler struct {
	repo       repository.CreditRepository
	openapiCli openapiv1.OpenAPIServiceClient
	l          logger.LoggerV1
}

func NewHandler(
	repo repository.CreditRepository,
	openapiCli openapiv1.OpenAPIServiceClient,
	l logger.LoggerV1,
) *Handler {
	return &Handler{
		repo:       repo,
		openapiCli: openapiCli,
		l:          l,
	}
}

// RegisterRoutes 注册易支付路由
func (h *Handler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/mapi")
	{
		// 发起支付（兼容易支付规范）
		g.GET("/submit.php", h.Submit)
		g.POST("/submit.php", h.Submit)

		// 查询订单
		g.GET("/query.php", h.Query)
		g.POST("/query.php", h.Query)

		// 支付页面（用户确认支付）
		g.GET("/pay/:trade_no", h.PayPage)
		g.POST("/pay/:trade_no/confirm", h.PayConfirm)
	}
}

// Submit 发起支付
// 兼容易支付规范：接收form参数，验证签名，创建订单，跳转支付页面
func (h *Handler) Submit(ctx *gin.Context) {
	var req SubmitRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusOK, SubmitResponse{Code: -1, Msg: "参数错误: " + err.Error()})
		return
	}

	// 默认支付类型
	if req.Type == "" {
		req.Type = PayTypeCredit
	}

	// 验证商户（通过openapi服务）
	_, secret, err := h.verifyMerchant(ctx.Request.Context(), req.Pid)
	if err != nil {
		ctx.JSON(http.StatusOK, SubmitResponse{Code: -1, Msg: err.Error()})
		return
	}

	// 验证签名
	if !VerifySign(req.ToMap(), secret, req.Sign) {
		ctx.JSON(http.StatusOK, SubmitResponse{Code: -1, Msg: ErrInvalidSign.Error()})
		return
	}

	// 解析金额
	money, err := strconv.ParseInt(req.Money, 10, 64)
	if err != nil || money <= 0 {
		ctx.JSON(http.StatusOK, SubmitResponse{Code: -1, Msg: "金额格式错误"})
		return
	}

	// 检查是否已存在相同订单（幂等处理）
	existOrder, err := h.repo.GetEpayOrderByOutTradeNo(ctx.Request.Context(), req.Pid, req.OutTradeNo)
	if err == nil {
		// 订单已存在
		if existOrder.Status == domain.EpayOrderStatusSuccess {
			ctx.JSON(http.StatusOK, SubmitResponse{Code: -1, Msg: ErrOrderPaid.Error()})
			return
		}
		// 返回已有订单的支付页面
		h.redirectToPayPage(ctx, existOrder.TradeNo, req.ReturnURL)
		return
	}
	if !errors.Is(err, repository.ErrEpayOrderNotFound) {
		h.l.Error("查询订单失败", logger.Error(err))
		ctx.JSON(http.StatusOK, SubmitResponse{Code: -1, Msg: "系统错误"})
		return
	}

	// 生成平台订单号
	tradeNo := h.generateTradeNo()

	// 创建订单
	order := domain.EpayOrder{
		TradeNo:    tradeNo,
		OutTradeNo: req.OutTradeNo,
		AppId:      req.Pid,
		Uid:        req.Uid,
		Type:       req.Type,
		Name:       req.Name,
		Money:      money,
		Status:     domain.EpayOrderStatusWait,
		NotifyURL:  req.NotifyURL,
		ReturnURL:  req.ReturnURL,
		Param:      req.Param,
	}

	_, err = h.repo.CreateEpayOrder(ctx.Request.Context(), order)
	if err != nil {
		h.l.Error("创建订单失败", logger.Error(err))
		ctx.JSON(http.StatusOK, SubmitResponse{Code: -1, Msg: "创建订单失败"})
		return
	}

	// 跳转到支付页面
	h.redirectToPayPage(ctx, tradeNo, req.ReturnURL)
}

// Query 查询订单
func (h *Handler) Query(ctx *gin.Context) {
	var req QueryRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusOK, QueryResponse{Code: -1, Msg: "参数错误"})
		return
	}

	// 验证商户
	_, secret, err := h.verifyMerchant(ctx.Request.Context(), req.Pid)
	if err != nil {
		ctx.JSON(http.StatusOK, QueryResponse{Code: -1, Msg: err.Error()})
		return
	}

	// 验证签名
	if !VerifySign(req.ToMap(), secret, req.Sign) {
		ctx.JSON(http.StatusOK, QueryResponse{Code: -1, Msg: ErrInvalidSign.Error()})
		return
	}

	// 查询订单
	var order domain.EpayOrder
	if req.TradeNo != "" {
		order, err = h.repo.GetEpayOrderByTradeNo(ctx.Request.Context(), req.TradeNo)
	} else if req.OutTradeNo != "" {
		order, err = h.repo.GetEpayOrderByOutTradeNo(ctx.Request.Context(), req.Pid, req.OutTradeNo)
	} else {
		ctx.JSON(http.StatusOK, QueryResponse{Code: -1, Msg: "订单号不能为空"})
		return
	}

	if err != nil {
		if errors.Is(err, repository.ErrEpayOrderNotFound) {
			ctx.JSON(http.StatusOK, QueryResponse{Code: -1, Msg: ErrInvalidOrder.Error()})
			return
		}
		h.l.Error("查询订单失败", logger.Error(err))
		ctx.JSON(http.StatusOK, QueryResponse{Code: -1, Msg: "系统错误"})
		return
	}

	// 验证订单归属
	if order.AppId != req.Pid {
		ctx.JSON(http.StatusOK, QueryResponse{Code: -1, Msg: ErrInvalidOrder.Error()})
		return
	}

	ctx.JSON(http.StatusOK, QueryResponse{
		Code:        1,
		Msg:         "success",
		TradeNo:     order.TradeNo,
		OutTradeNo:  order.OutTradeNo,
		Type:        order.Type,
		Pid:         order.AppId,
		Name:        order.Name,
		Money:       strconv.FormatInt(order.Money, 10),
		TradeStatus: h.statusToTradeStatus(order.Status),
		Param:       order.Param,
		Addtime:     time.Unix(order.Ctime/1000, 0).Format("2006-01-02 15:04:05"),
		Endtime:     h.getEndTime(order),
	})
}

// PayPage 支付页面
func (h *Handler) PayPage(ctx *gin.Context) {
	tradeNo := ctx.Param("trade_no")
	if tradeNo == "" {
		ctx.String(http.StatusBadRequest, "订单号不能为空")
		return
	}

	order, err := h.repo.GetEpayOrderByTradeNo(ctx.Request.Context(), tradeNo)
	if err != nil {
		if errors.Is(err, repository.ErrEpayOrderNotFound) {
			ctx.String(http.StatusNotFound, "订单不存在")
			return
		}
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}

	// 检查订单状态
	if order.Status != domain.EpayOrderStatusWait {
		ctx.String(http.StatusBadRequest, "订单状态异常")
		return
	}

	// 获取用户余额
	balance, err := h.repo.GetAccount(ctx.Request.Context(), order.Uid)
	if err != nil && !errors.Is(err, repository.ErrAccountNotFound) {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}

	// 返回简单的HTML支付页面
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>积分支付</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 400px; margin: 50px auto; padding: 20px; }
        .card { background: #fff; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); padding: 20px; }
        .title { font-size: 18px; font-weight: bold; margin-bottom: 20px; text-align: center; }
        .info { margin: 10px 0; display: flex; justify-content: space-between; }
        .label { color: #666; }
        .value { font-weight: bold; }
        .amount { font-size: 24px; color: #ff6b00; text-align: center; margin: 20px 0; }
        .balance { text-align: center; color: #666; margin-bottom: 20px; }
        .btn { width: 100%%; padding: 12px; background: #ff6b00; color: #fff; border: none; border-radius: 4px; font-size: 16px; cursor: pointer; }
        .btn:hover { background: #e65c00; }
        .btn:disabled { background: #ccc; cursor: not-allowed; }
        .error { color: #f00; text-align: center; margin-top: 10px; display: none; }
    </style>
</head>
<body>
    <div class="card">
        <div class="title">%s</div>
        <div class="info"><span class="label">订单号</span><span class="value">%s</span></div>
        <div class="amount">%d 积分</div>
        <div class="balance">当前余额：%d 积分</div>
        <form action="/mapi/pay/%s/confirm" method="POST">
            <button type="submit" class="btn" %s>确认支付</button>
        </form>
        <div class="error" id="error">%s</div>
    </div>
</body>
</html>`,
		order.Name,
		order.OutTradeNo,
		order.Money,
		balance.Balance,
		tradeNo,
		func() string {
			if balance.Balance < order.Money {
				return "disabled"
			}
			return ""
		}(),
		func() string {
			if balance.Balance < order.Money {
				return "积分余额不足"
			}
			return ""
		}(),
	)

	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.String(http.StatusOK, html)
}

// PayConfirm 确认支付
func (h *Handler) PayConfirm(ctx *gin.Context) {
	tradeNo := ctx.Param("trade_no")
	if tradeNo == "" {
		ctx.String(http.StatusBadRequest, "订单号不能为空")
		return
	}

	order, err := h.repo.GetEpayOrderByTradeNo(ctx.Request.Context(), tradeNo)
	if err != nil {
		if errors.Is(err, repository.ErrEpayOrderNotFound) {
			ctx.String(http.StatusNotFound, "订单不存在")
			return
		}
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}

	// 检查订单状态
	if order.Status != domain.EpayOrderStatusWait {
		ctx.String(http.StatusBadRequest, "订单已处理")
		return
	}

	// 执行扣款
	_, err = h.repo.AddCredit(ctx.Request.Context(), order.Uid, "epay_pay", order.Id, -order.Money, fmt.Sprintf("支付订单: %s", order.Name))
	if err != nil {
		h.l.Error("扣款失败", logger.Error(err), logger.String("trade_no", tradeNo))
		// 可能是余额不足
		ctx.String(http.StatusBadRequest, "支付失败: 余额不足")
		return
	}

	// 更新订单状态
	if err = h.repo.UpdateEpayOrderStatus(ctx.Request.Context(), order.Id, domain.EpayOrderStatusSuccess); err != nil {
		h.l.Error("更新订单状态失败", logger.Error(err), logger.String("trade_no", tradeNo))
		// 继续处理，不影响用户
	}

	// 异步发送通知（通过goroutine）
	go h.sendNotify(order)

	// 跳转到return_url
	if order.ReturnURL != "" {
		returnURL := h.buildReturnURL(order)
		ctx.Redirect(http.StatusFound, returnURL)
		return
	}

	ctx.String(http.StatusOK, "支付成功")
}

// verifyMerchant 验证商户
func (h *Handler) verifyMerchant(ctx context.Context, pid string) (*openapiv1.App, string, error) {
	resp, err := h.openapiCli.GetAppSecret(ctx, &openapiv1.GetAppSecretRequest{AppId: pid})
	if err != nil {
		h.l.Error("获取商户信息失败", logger.Error(err), logger.String("pid", pid))
		return nil, "", ErrInvalidPid
	}
	if resp.GetApp() == nil {
		return nil, "", ErrInvalidPid
	}
	if resp.GetApp().GetStatus() != 1 { // 1表示启用
		return nil, "", errors.New("商户已禁用")
	}
	return resp.GetApp(), resp.GetSecret(), nil
}

// generateTradeNo 生成平台订单号
func (h *Handler) generateTradeNo() string {
	return fmt.Sprintf("EP%d%d", time.Now().UnixNano()/1000000, time.Now().UnixNano()%10000)
}

// redirectToPayPage 重定向到支付页面
func (h *Handler) redirectToPayPage(ctx *gin.Context, tradeNo, returnURL string) {
	payURL := fmt.Sprintf("/mapi/pay/%s", tradeNo)
	ctx.Redirect(http.StatusFound, payURL)
}

// statusToTradeStatus 状态转换
func (h *Handler) statusToTradeStatus(status domain.EpayOrderStatus) string {
	switch status {
	case domain.EpayOrderStatusWait:
		return TradeStatusWait
	case domain.EpayOrderStatusSuccess, domain.EpayOrderStatusNotified:
		return TradeStatusSuccess
	case domain.EpayOrderStatusClosed:
		return TradeStatusClosed
	default:
		return TradeStatusWait
	}
}

// getEndTime 获取完成时间
func (h *Handler) getEndTime(order domain.EpayOrder) string {
	if order.Status == domain.EpayOrderStatusSuccess || order.Status == domain.EpayOrderStatusNotified {
		return time.Unix(order.Utime/1000, 0).Format("2006-01-02 15:04:05")
	}
	return ""
}

// sendNotify 发送异步通知
func (h *Handler) sendNotify(order domain.EpayOrder) {
	if order.NotifyURL == "" {
		return
	}

	// 获取商户密钥
	_, secret, err := h.verifyMerchant(context.Background(), order.AppId)
	if err != nil {
		h.l.Error("获取商户信息失败", logger.Error(err))
		return
	}

	// 构建通知参数
	params := map[string]string{
		"pid":          order.AppId,
		"trade_no":     order.TradeNo,
		"out_trade_no": order.OutTradeNo,
		"type":         order.Type,
		"name":         order.Name,
		"money":        strconv.FormatInt(order.Money, 10),
		"trade_status": TradeStatusSuccess,
		"param":        order.Param,
		"sign_type":    "MD5",
	}
	params["sign"] = Sign(params, secret)

	// 发送通知（最多重试5次）
	for i := 0; i < 5; i++ {
		success := h.doNotify(order.NotifyURL, params)
		now := time.Now().UnixMilli()

		// 更新通知信息
		_ = h.repo.UpdateEpayOrderNotify(context.Background(), order.Id, i+1, now)

		if success {
			// 更新状态为已通知
			_ = h.repo.UpdateEpayOrderStatus(context.Background(), order.Id, domain.EpayOrderStatusNotified)
			return
		}

		// 等待后重试，时间逐渐增加
		time.Sleep(time.Duration(i+1) * 30 * time.Second)
	}

	h.l.Error("通知失败，已达最大重试次数",
		logger.String("trade_no", order.TradeNo),
		logger.String("notify_url", order.NotifyURL))
}

// doNotify 执行通知
func (h *Handler) doNotify(notifyURL string, params map[string]string) bool {
	// 构建POST表单
	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}

	resp, err := http.PostForm(notifyURL, form)
	if err != nil {
		h.l.Error("发送通知失败", logger.Error(err), logger.String("url", notifyURL))
		return false
	}
	defer resp.Body.Close()

	// 读取响应，商户返回"success"表示成功
	buf := make([]byte, 64)
	n, _ := resp.Body.Read(buf)
	body := strings.TrimSpace(strings.ToLower(string(buf[:n])))

	return body == "success"
}

// buildReturnURL 构建同步跳转地址
func (h *Handler) buildReturnURL(order domain.EpayOrder) string {
	// 获取商户密钥
	_, secret, err := h.verifyMerchant(context.Background(), order.AppId)
	if err != nil {
		return order.ReturnURL
	}

	// 构建参数
	params := map[string]string{
		"pid":          order.AppId,
		"trade_no":     order.TradeNo,
		"out_trade_no": order.OutTradeNo,
		"type":         order.Type,
		"name":         order.Name,
		"money":        strconv.FormatInt(order.Money, 10),
		"trade_status": TradeStatusSuccess,
		"param":        order.Param,
		"sign_type":    "MD5",
	}
	params["sign"] = Sign(params, secret)

	// 构建URL
	u, err := url.Parse(order.ReturnURL)
	if err != nil {
		return order.ReturnURL
	}

	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	return u.String()
}
