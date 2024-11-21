package handler

import (
	"github.com/gin-gonic/gin"
	articlev1 "webook/api/proto/gen/article/v1"
	rewardv1 "webook/api/proto/gen/reward/v1"
	"webook/internal/handler/jwt"
	"webook/pkg/ginx"
	"webook/pkg/logger"
)

type RewardHandler struct {
	client    rewardv1.RewardServiceClient
	artClient articlev1.ArticleServiceClient
	l         logger.LoggerV1
}

func NewRewardHandler(client rewardv1.RewardServiceClient, artClient articlev1.ArticleServiceClient) *RewardHandler {
	return &RewardHandler{client: client, artClient: artClient}
}

func (h *RewardHandler) RegisterRoutes(server *gin.Engine) {
	rg := server.Group("/reward")
	rg.POST("/detail", ginx.WrapBodyAndToken[GetRewardReq](h.l, h.GetReward))
}

type GetRewardReq struct {
	Rid int64
}

func (h *RewardHandler) GetReward(ctx *gin.Context, req GetRewardReq, uc jwt.UserClaims) (ginx.Result, error) {
	resp, err := h.client.GetReward(ctx.Request.Context(), &rewardv1.GetRewardRequest{
		Rid: req.Rid,
		Uid: uc.UserId,
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		// 暂时只需要状态
		Data: resp.Status.String(),
	}, nil
}

type RewardArticleReq struct {
	Aid int64 `json:"aid"`
	Amt int64 `json:"amt"`
}
