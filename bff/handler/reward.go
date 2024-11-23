package handler

import (
	"context"
	"github.com/gin-gonic/gin"
	"time"
	articlev1 "webook/api/proto/gen/article/v1"
	rewardv1 "webook/api/proto/gen/reward/v1"
	"webook/bff/handler/jwt"
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

// GetRewardV1 前端传过来一个超长的超时时间，例如说 10s
// 后端进行轮询
// 可能引来巨大的性能问题
// 真正优雅的还是前端来轮询
// 考虑改用 stream：sse
func (h *RewardHandler) GetRewardV1(ctx *gin.Context, req GetRewardReq, uc jwt.UserClaims) (ginx.Result, error) {
	// 前端不愿意轮询查询支付状态的话
	for {
		newCtx, cancel := context.WithTimeout(ctx.Request.Context(), time.Second)
		resp, err := h.client.GetReward(newCtx, &rewardv1.GetRewardRequest{
			Rid: req.Rid,
			Uid: uc.UserId,
		})
		cancel()
		if err != nil {
			return ginx.Result{
				Code: 5,
				Msg:  "系统错误",
			}, err
		}
		if resp.GetStatus() == 1 {
			continue
		}
		return ginx.Result{
			// 暂时只需要状态
			Data: resp.Status.String(),
		}, nil
	}
}

type RewardArticleReq struct {
	Aid int64 `json:"aid"`
	Amt int64 `json:"amt"`
}
