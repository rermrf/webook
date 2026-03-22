package handler

import (
	creditv1 "webook/api/proto/gen/credit/v1"
	"webook/bff/handler/jwt"
	"webook/pkg/ginx"
	"webook/pkg/logger"

	"github.com/gin-gonic/gin"
)

type CreditHandler struct {
	client creditv1.CreditServiceClient
	l      logger.LoggerV1
}

func NewCreditHandler(client creditv1.CreditServiceClient, l logger.LoggerV1) *CreditHandler {
	return &CreditHandler{
		client: client,
		l:      l,
	}
}

func (h *CreditHandler) RegisterRoutes(server *gin.Engine) {
	rg := server.Group("/credit")
	rg.GET("/balance", ginx.WrapClaims[jwt.UserClaims](h.l, h.GetBalance))
	rg.POST("/flows", ginx.WrapBodyAndToken[GetFlowsReq, jwt.UserClaims](h.l, h.GetFlows))
	rg.POST("/recharge", ginx.WrapBodyAndToken[PreRechargeReq, jwt.UserClaims](h.l, h.PreRecharge))
	rg.POST("/recharge/status", ginx.WrapBodyAndToken[GetRechargeReq, jwt.UserClaims](h.l, h.GetRecharge))
	rg.POST("/reward", ginx.WrapBodyAndToken[RewardCreditReq, jwt.UserClaims](h.l, h.RewardCredit))
	rg.POST("/reward/detail", ginx.WrapBodyAndToken[GetCreditRewardReq, jwt.UserClaims](h.l, h.GetCreditReward))
	rg.GET("/daily-status", ginx.WrapClaims[jwt.UserClaims](h.l, h.GetDailyStatus))
	rg.POST("/sign-in", ginx.WrapClaims[jwt.UserClaims](h.l, h.SignIn))
}

// GetBalance 查询积分余额
func (h *CreditHandler) GetBalance(ctx *gin.Context, uc jwt.UserClaims) (ginx.Result, error) {
	resp, err := h.client.GetBalance(ctx.Request.Context(), &creditv1.GetBalanceRequest{
		Uid: uc.UserId,
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Data: resp.Balance,
	}, nil
}

type GetFlowsReq struct {
	Offset int64 `json:"offset"`
	Limit  int64 `json:"limit"`
}

type CreditFlowVO struct {
	Id          int64  `json:"id"`
	Biz         string `json:"biz"`
	BizId       int64  `json:"biz_id"`
	ChangeAmt   int64  `json:"change_amt"`
	Balance     int64  `json:"balance"`
	Description string `json:"description"`
	Ctime       int64  `json:"ctime"`
}

// GetFlows 查询积分流水
func (h *CreditHandler) GetFlows(ctx *gin.Context, req GetFlowsReq, uc jwt.UserClaims) (ginx.Result, error) {
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	resp, err := h.client.GetFlows(ctx.Request.Context(), &creditv1.GetFlowsRequest{
		Uid:    uc.UserId,
		Offset: req.Offset,
		Limit:  req.Limit,
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}

	flows := make([]CreditFlowVO, 0, len(resp.Flows))
	for _, f := range resp.Flows {
		flows = append(flows, CreditFlowVO{
			Id:          f.Id,
			Biz:         f.Biz,
			BizId:       f.BizId,
			ChangeAmt:   f.ChangeAmt,
			Balance:     f.Balance,
			Description: f.Description,
			Ctime:       f.Ctime,
		})
	}
	return ginx.Result{
		Data: flows,
	}, nil
}

type PreRechargeReq struct {
	CreditAmt int64 `json:"credit_amt"` // 要购买的积分数量
}

type PreRechargeResp struct {
	RechargeId int64  `json:"recharge_id"`
	CodeUrl    string `json:"code_url"`
	PaymentAmt int64  `json:"payment_amt"` // 需要支付的金额（分）
}

// PreRecharge 发起积分充值
func (h *CreditHandler) PreRecharge(ctx *gin.Context, req PreRechargeReq, uc jwt.UserClaims) (ginx.Result, error) {
	if req.CreditAmt <= 0 {
		return ginx.Result{
			Code: 4,
			Msg:  "积分数量必须大于0",
		}, nil
	}

	resp, err := h.client.PreRecharge(ctx.Request.Context(), &creditv1.PreRechargeRequest{
		Uid:       uc.UserId,
		CreditAmt: req.CreditAmt,
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Data: PreRechargeResp{
			RechargeId: resp.RechargeId,
			CodeUrl:    resp.CodeUrl,
			PaymentAmt: resp.PaymentAmt,
		},
	}, nil
}

type GetRechargeReq struct {
	RechargeId int64 `json:"recharge_id"`
}

// GetRecharge 查询充值状态
func (h *CreditHandler) GetRecharge(ctx *gin.Context, req GetRechargeReq, uc jwt.UserClaims) (ginx.Result, error) {
	resp, err := h.client.GetRecharge(ctx.Request.Context(), &creditv1.GetRechargeRequest{
		RechargeId: req.RechargeId,
		Uid:        uc.UserId,
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Data: map[string]interface{}{
			"status":     resp.Status.String(),
			"credit_amt": resp.CreditAmt,
		},
	}, nil
}

type RewardCreditReq struct {
	TargetUid int64  `json:"target_uid"` // 被打赏者
	Biz       string `json:"biz"`        // 业务类型
	BizId     int64  `json:"biz_id"`     // 业务ID
	Amt       int64  `json:"amt"`        // 打赏积分数量
}

// RewardCredit 积分打赏
func (h *CreditHandler) RewardCredit(ctx *gin.Context, req RewardCreditReq, uc jwt.UserClaims) (ginx.Result, error) {
	if req.Amt <= 0 {
		return ginx.Result{
			Code: 4,
			Msg:  "打赏积分必须大于0",
		}, nil
	}

	if uc.UserId == req.TargetUid {
		return ginx.Result{
			Code: 4,
			Msg:  "不能给自己打赏",
		}, nil
	}

	resp, err := h.client.RewardCredit(ctx.Request.Context(), &creditv1.RewardCreditRequest{
		Uid:       uc.UserId,
		TargetUid: req.TargetUid,
		Biz:       req.Biz,
		BizId:     req.BizId,
		Amt:       req.Amt,
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}

	if !resp.Success {
		return ginx.Result{
			Code: 4,
			Msg:  resp.Message,
		}, nil
	}

	return ginx.Result{
		Data: map[string]interface{}{
			"reward_id": resp.RewardId,
			"success":   resp.Success,
		},
	}, nil
}

type GetCreditRewardReq struct {
	RewardId int64 `json:"reward_id"`
}

// GetCreditReward 查询打赏详情
func (h *CreditHandler) GetCreditReward(ctx *gin.Context, req GetCreditRewardReq, uc jwt.UserClaims) (ginx.Result, error) {
	resp, err := h.client.GetCreditReward(ctx.Request.Context(), &creditv1.GetCreditRewardRequest{
		RewardId: req.RewardId,
		Uid:      uc.UserId,
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Data: map[string]interface{}{
			"status": resp.Status.String(),
			"amt":    resp.Amt,
		},
	}, nil
}

type DailyStatusVO struct {
	Biz         string `json:"biz"`
	EarnedCount int64  `json:"earned_count"`
	EarnedAmt   int64  `json:"earned_amt"`
	DailyLimit  int64  `json:"daily_limit"`
	Remaining   int64  `json:"remaining"`
}

// GetDailyStatus 查询每日积分获取状态
func (h *CreditHandler) GetDailyStatus(ctx *gin.Context, uc jwt.UserClaims) (ginx.Result, error) {
	biz := ctx.Query("biz")
	resp, err := h.client.GetDailyStatus(ctx.Request.Context(), &creditv1.GetDailyStatusRequest{
		Uid: uc.UserId,
		Biz: biz,
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}

	statuses := make([]DailyStatusVO, 0, len(resp.Statuses))
	for _, s := range resp.Statuses {
		statuses = append(statuses, DailyStatusVO{
			Biz:         s.Biz,
			EarnedCount: s.EarnedCount,
			EarnedAmt:   s.EarnedAmt,
			DailyLimit:  s.DailyLimit,
			Remaining:   s.Remaining,
		})
	}
	return ginx.Result{
		Data: statuses,
	}, nil
}

// SignIn 每日签到，调用 EarnCredit 获取积分
func (h *CreditHandler) SignIn(ctx *gin.Context, uc jwt.UserClaims) (ginx.Result, error) {
	resp, err := h.client.EarnCredit(ctx.Request.Context(), &creditv1.EarnCreditRequest{
		Uid: uc.UserId,
		Biz: "daily_signin",
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}

	if !resp.Success {
		return ginx.Result{
			Code: 4,
			Msg:  resp.Message,
		}, nil
	}

	return ginx.Result{
		Data: map[string]interface{}{
			"earned_amt":  resp.EarnedAmt,
			"new_balance": resp.NewBalance,
		},
	}, nil
}
