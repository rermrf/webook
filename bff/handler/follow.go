package handler

import (
	"github.com/gin-gonic/gin"
	followv1 "webook/api/proto/gen/follow/v1"
	userv1 "webook/api/proto/gen/user/v1"
	ijwt "webook/bff/handler/jwt"
	"webook/pkg/ginx"
	"webook/pkg/logger"
)

type FollowHandler struct {
	svc     followv1.FollowServiceClient
	userSvc userv1.UserServiceClient
	l       logger.LoggerV1
}

func (h *FollowHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/follow")
	ug.POST("/follow", ginx.WrapBodyAndToken(h.l, h.Follow))
}

type FolloweeRequest struct {
	Followee int64 `json:"followee"`
}

type FollowerRequest struct {
	Follower int64 `json:"follower"`
}

func (h *FollowHandler) Follow(ctx *gin.Context, req FolloweeRequest, uc ijwt.UserClaims) (ginx.Result, error) {
	_, err := h.svc.Follow(ctx.Request.Context(), &followv1.FollowRequest{
		Followee: req.Followee,
		Follower: uc.UserId,
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, nil
	}
	return ginx.Result{
		Code: 2,
		Msg:  "OK",
	}, nil
}

func (h *FollowHandler) CanelFollow(ctx *gin.Context, req FolloweeRequest, uc ijwt.UserClaims) (ginx.Result, error) {
	_, err := h.svc.CancelFollow(ctx.Request.Context(), &followv1.CancelFollowRequest{
		Followee: req.Followee,
		Follower: uc.UserId,
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, nil
	}
	return ginx.Result{
		Code: 2,
		Msg:  "OK",
	}, nil
}

func (h *FollowHandler) GetFollowee(ctx *gin.Context, req FollowerRequest, uc ijwt.UserClaims) (ginx.Result, error) {
	var resp *followv1.GetFolloweeResponse
	var err error
	if req.Follower == 0 {
		resp, err = h.svc.GetFollowee(ctx.Request.Context(), &followv1.GetFolloweeRequest{
			Follower: uc.UserId,
		})
	} else {
		resp, err = h.svc.GetFollowee(ctx.Request.Context(), &followv1.GetFolloweeRequest{
			Follower: req.Follower,
		})
	}
	h.userSvc
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, nil
	}
	resp.GetFollowRelations()
}

// GetFollowStatic 传入一个用户id，获取他的关注人数和粉丝数，如果这个id为空则获取自己的
func (h *FollowHandler) GetFollowStatic(ctx *gin.Context, req FolloweeRequest, uc ijwt.UserClaims) (ginx.Result, error) {
	type Response struct {
		// 关注的人数
		Followees int64 `json:"followees"`
		// 粉丝数
		Followers int64 `json:"followers"`
	}
	var resp *followv1.GetFollowStaticResponse
	var err error
	if req.Followee == 0 {
		resp, err = h.svc.GetFollowStatic(ctx.Request.Context(), &followv1.GetFollowStaticRequest{
			Followee: req.Followee,
		})
	} else {
		resp, err = h.svc.GetFollowStatic(ctx.Request.Context(), &followv1.GetFollowStaticRequest{
			Followee: uc.UserId,
		})
	}

	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, nil
	}
	return ginx.Result{
		Code: 2,
		Msg:  "OK",
		Data: Response{
			Followees: resp.GetFollowStatic().GetFollowees(),
			Followers: resp.GetFollowStatic().GetFollowers(),
		},
	}, nil
}
