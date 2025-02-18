package handler

import (
	"github.com/gin-gonic/gin"
	"time"
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

func NewFollowHandler(svc followv1.FollowServiceClient, userSvc userv1.UserServiceClient, l logger.LoggerV1) *FollowHandler {
	return &FollowHandler{svc: svc, userSvc: userSvc, l: l}
}

func (h *FollowHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/follow")
	ug.POST("/follow", ginx.WrapBodyAndToken(h.l, h.Follow))
	ug.POST("/cancel", ginx.WrapBodyAndToken(h.l, h.CanelFollow))
	ug.GET("/followee", ginx.WrapBodyAndToken(h.l, h.GetFollowee))
	ug.GET("/follower", ginx.WrapBodyAndToken(h.l, h.GetFollower))
	ug.GET("/static", ginx.WrapBodyAndToken(h.l, h.GetFollowStatic))
}

type FolloweeRequest struct {
	Followee int64 `json:"followee"`
}

type FollowerRequest struct {
	Follower int64 `json:"follower"`
}

// Follow 关注接口
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

// CanelFollow 取消关注
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

// GetFollowee 获取某人的关注列表
// TODO: 在 user 模块添加批量查询接口，提升性能
func (h *FollowHandler) GetFollowee(ctx *gin.Context, req FollowerRequest, uc ijwt.UserClaims) (ginx.Result, error) {
	var resp *followv1.GetFolloweeResponse
	var err error
	if req.Follower == 0 {
		resp, err = h.svc.GetFollowee(ctx.Request.Context(), &followv1.GetFolloweeRequest{
			Follower: uc.UserId,
			Offset:   0,
			Limit:    1000,
		})
	} else {
		resp, err = h.svc.GetFollowee(ctx.Request.Context(), &followv1.GetFolloweeRequest{
			Follower: req.Follower,
			Offset:   0,
			Limit:    1000,
		})
	}
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, nil
	}
	var res []Profile
	for _, relation := range resp.GetFollowRelations() {
		pResp, err := h.userSvc.Profile(ctx, &userv1.ProfileRequest{
			Id: relation.Followee,
		})
		if err != nil {
			continue
		}
		res = append(res, Profile{
			Id:       pResp.GetUser().GetId(),
			Email:    pResp.GetUser().GetEmail(),
			Phone:    pResp.GetUser().GetPhone(),
			Nickname: pResp.GetUser().GetNickName(),
			AboutMe:  pResp.GetUser().GetAboutMe(),
			Birthday: pResp.GetUser().GetBirthday().AsTime().Format(time.DateOnly),
			Ctime:    pResp.GetUser().GetCtime().AsTime().Format(time.DateOnly),
		})
	}
	return ginx.Result{
		Code: 2,
		Msg:  "OK",
		Data: res,
	}, nil
}

// GetFollower 获取某人的粉丝列表
// TODO: 在 user 模块添加批量查询接口，提升性能
func (h *FollowHandler) GetFollower(ctx *gin.Context, req FollowerRequest, uc ijwt.UserClaims) (ginx.Result, error) {
	var resp *followv1.GetFollowerResponse
	var err error
	if req.Follower == 0 {
		resp, err = h.svc.GetFollower(ctx.Request.Context(), &followv1.GetFollowerRequest{
			Followee: uc.UserId,
			Offset:   0,
			Limit:    1000,
		})
	} else {
		resp, err = h.svc.GetFollower(ctx.Request.Context(), &followv1.GetFollowerRequest{
			Followee: req.Follower,
			Offset:   0,
			Limit:    1000,
		})
	}
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, nil
	}
	var res []Profile
	for _, relation := range resp.GetFollowRelations() {
		pResp, err := h.userSvc.Profile(ctx, &userv1.ProfileRequest{
			Id: relation.Follower,
		})
		if err != nil {
			continue
		}
		res = append(res, Profile{
			Id:       pResp.GetUser().GetId(),
			Email:    pResp.GetUser().GetEmail(),
			Phone:    pResp.GetUser().GetPhone(),
			Nickname: pResp.GetUser().GetNickName(),
			AboutMe:  pResp.GetUser().GetAboutMe(),
			Birthday: pResp.GetUser().GetBirthday().AsTime().Format(time.DateOnly),
			Ctime:    pResp.GetUser().GetCtime().AsTime().Format(time.DateOnly),
		})
	}
	return ginx.Result{
		Code: 2,
		Msg:  "OK",
		Data: res,
	}, nil
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
			Followee: uc.UserId,
		})
	} else {
		resp, err = h.svc.GetFollowStatic(ctx.Request.Context(), &followv1.GetFollowStaticRequest{
			Followee: req.Followee,
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
