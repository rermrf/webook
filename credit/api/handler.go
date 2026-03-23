package api

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	openapiv1 "webook/api/proto/gen/openapi/v1"
	"webook/credit/service"
	"webook/pkg/logger"
)

// Handler 积分开放API HTTP处理器
type Handler struct {
	openapiCli openapiv1.OpenAPIServiceClient
	creditSvc  service.CreditService
	l          logger.LoggerV1
}

func NewHandler(
	openapiCli openapiv1.OpenAPIServiceClient,
	creditSvc service.CreditService,
	l logger.LoggerV1,
) *Handler {
	return &Handler{
		openapiCli: openapiCli,
		creditSvc:  creditSvc,
		l:          l,
	}
}

// RegisterRoutes 注册积分开放API路由
func (h *Handler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/openapi/v1")
	{
		// 需要签名验证的接口
		g.POST("/credit/balance", h.GetBalance)      // 查询余额
		g.POST("/credit/deduct", h.DeductCredit)     // 扣减积分
		g.POST("/credit/transfer", h.TransferCredit) // 转账积分
	}
}

// ======== 开放API接口（需要签名验证）========

// GetBalanceReq 查询余额请求
type GetBalanceReq struct {
	AppId     string `json:"app_id" binding:"required"`
	Timestamp int64  `json:"timestamp" binding:"required"`
	Nonce     string `json:"nonce" binding:"required"`
	Sign      string `json:"sign" binding:"required"`
	Uid       int64  `json:"uid" binding:"required"`
}

// GetBalance 查询用户积分余额
func (h *Handler) GetBalance(ctx *gin.Context) {
	var req GetBalanceReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, Result{Code: 400, Msg: "参数错误"})
		return
	}

	// 验证请求
	app, err := h.verifyRequest(ctx.Request.Context(), SignParams{
		AppId:     req.AppId,
		Timestamp: req.Timestamp,
		Nonce:     req.Nonce,
		Body:      strconv.FormatInt(req.Uid, 10),
	}, req.Sign, ctx.ClientIP())
	if err != nil {
		h.handleError(ctx, err)
		return
	}
	_ = app // 可用于记录日志

	balance, err := h.creditSvc.GetBalance(ctx.Request.Context(), req.Uid)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{Code: 500, Msg: "查询失败"})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "success",
		Data: gin.H{"balance": balance},
	})
}

// DeductCreditReq 扣减积分请求
type DeductCreditReq struct {
	AppId       string `json:"app_id" binding:"required"`
	Timestamp   int64  `json:"timestamp" binding:"required"`
	Nonce       string `json:"nonce" binding:"required"`
	Sign        string `json:"sign" binding:"required"`
	OutTradeNo  string `json:"out_trade_no" binding:"required"` // 外部交易号
	Uid         int64  `json:"uid" binding:"required"`
	Amount      int64  `json:"amount" binding:"required,gt=0"`
	Description string `json:"description"`
}

// DeductCredit 扣减积分
func (h *Handler) DeductCredit(ctx *gin.Context) {
	var req DeductCreditReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, Result{Code: 400, Msg: "参数错误"})
		return
	}

	// 验证请求
	_, err := h.verifyRequest(ctx.Request.Context(), SignParams{
		AppId:     req.AppId,
		Timestamp: req.Timestamp,
		Nonce:     req.Nonce,
		Body:      req.OutTradeNo + strconv.FormatInt(req.Uid, 10) + strconv.FormatInt(req.Amount, 10),
	}, req.Sign, ctx.ClientIP())
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	// 生成交易号
	tradeNo := GenerateTradeNo(req.AppId)

	// 执行扣减
	err = h.creditSvc.DeductCredit(ctx.Request.Context(), req.Uid, req.Amount, "openapi_deduct", tradeNo, req.Description)
	if err != nil {
		if err == service.ErrInsufficientBalance {
			ctx.JSON(http.StatusBadRequest, Result{Code: 400, Msg: "积分余额不足"})
			return
		}
		h.l.Error("扣减积分失败", logger.Error(err))
		ctx.JSON(http.StatusInternalServerError, Result{Code: 500, Msg: "扣减失败"})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "success",
		Data: gin.H{
			"trade_no":     tradeNo,
			"out_trade_no": req.OutTradeNo,
		},
	})
}

// TransferCreditReq 转账请求
type TransferCreditReq struct {
	AppId       string `json:"app_id" binding:"required"`
	Timestamp   int64  `json:"timestamp" binding:"required"`
	Nonce       string `json:"nonce" binding:"required"`
	Sign        string `json:"sign" binding:"required"`
	OutTradeNo  string `json:"out_trade_no" binding:"required"`
	FromUid     int64  `json:"from_uid" binding:"required"`
	ToUid       int64  `json:"to_uid" binding:"required"`
	Amount      int64  `json:"amount" binding:"required,gt=0"`
	Description string `json:"description"`
}

// TransferCredit 转账积分
func (h *Handler) TransferCredit(ctx *gin.Context) {
	var req TransferCreditReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, Result{Code: 400, Msg: "参数错误"})
		return
	}

	// 验证请求
	_, err := h.verifyRequest(ctx.Request.Context(), SignParams{
		AppId:     req.AppId,
		Timestamp: req.Timestamp,
		Nonce:     req.Nonce,
		Body:      req.OutTradeNo + strconv.FormatInt(req.FromUid, 10) + strconv.FormatInt(req.ToUid, 10) + strconv.FormatInt(req.Amount, 10),
	}, req.Sign, ctx.ClientIP())
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	// 生成交易号
	tradeNo := GenerateTradeNo(req.AppId)

	// 执行转账
	err = h.creditSvc.Transfer(ctx.Request.Context(), req.FromUid, req.ToUid, req.Amount, req.Description)
	if err != nil {
		if err == service.ErrInsufficientBalance {
			ctx.JSON(http.StatusBadRequest, Result{Code: 400, Msg: "积分余额不足"})
			return
		}
		if err == service.ErrSelfTransfer {
			ctx.JSON(http.StatusBadRequest, Result{Code: 400, Msg: "不能给自己转账"})
			return
		}
		h.l.Error("转账积分失败", logger.Error(err))
		ctx.JSON(http.StatusInternalServerError, Result{Code: 500, Msg: "转账失败"})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "success",
		Data: gin.H{
			"trade_no":     tradeNo,
			"out_trade_no": req.OutTradeNo,
		},
	})
}

// verifyRequest 验证请求（应用状态、IP白名单、时间戳、签名）
func (h *Handler) verifyRequest(ctx context.Context, params SignParams, sign, clientIP string) (*openapiv1.App, error) {
	// 调用 openapi 服务验证签名
	// VerifySignature 会验证：应用状态、时间戳、IP白名单、签名
	resp, err := h.openapiCli.VerifySignature(ctx, &openapiv1.VerifySignatureRequest{
		AppId:     params.AppId,
		Timestamp: params.Timestamp,
		Nonce:     params.Nonce,
		Body:      params.Body,
		Sign:      sign,
		ClientIp:  clientIP,
	})
	if err != nil {
		h.l.Error("签名验证失败", logger.Error(err), logger.String("app_id", params.AppId))
		return nil, ErrInvalidSign
	}

	return resp.GetApp(), nil
}

// handleError 统一错误处理
func (h *Handler) handleError(ctx *gin.Context, err error) {
	switch err {
	case ErrAppNotFound:
		ctx.JSON(http.StatusUnauthorized, Result{Code: 401, Msg: "应用不存在"})
	case ErrAppDisabled:
		ctx.JSON(http.StatusForbidden, Result{Code: 403, Msg: "应用已禁用"})
	case ErrInvalidSign:
		ctx.JSON(http.StatusUnauthorized, Result{Code: 401, Msg: "签名错误"})
	case ErrTimestampExpired:
		ctx.JSON(http.StatusUnauthorized, Result{Code: 401, Msg: "请求已过期"})
	case ErrInsufficientBalance:
		ctx.JSON(http.StatusBadRequest, Result{Code: 400, Msg: "积分余额不足"})
	case ErrInvalidIPAddress:
		ctx.JSON(http.StatusForbidden, Result{Code: 403, Msg: "IP地址不在白名单中"})
	default:
		h.l.Error("开放API错误", logger.Error(err))
		ctx.JSON(http.StatusInternalServerError, Result{Code: 500, Msg: "服务器错误"})
	}
}
