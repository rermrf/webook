package handler

import (
	"context"
	"github.com/gin-gonic/gin"
	"time"
	followv1 "webook/api/proto/gen/follow/v1"
	userv1 "webook/api/proto/gen/user/v1"
	"webook/bff/events"
	ijwt "webook/bff/handler/jwt"
	"webook/pkg/ginx"
	"webook/pkg/logger"
)

type FollowHandler struct {
	svc            followv1.FollowServiceClient
	userSvc        userv1.UserServiceClient
	notifyProducer events.NotificationProducer
	l              logger.LoggerV1
}

func NewFollowHandler(svc followv1.FollowServiceClient, userSvc userv1.UserServiceClient, notifyProducer events.NotificationProducer, l logger.LoggerV1) *FollowHandler {
	return &FollowHandler{svc: svc, userSvc: userSvc, notifyProducer: notifyProducer, l: l}
}

func (h *FollowHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/follow")
	ug.POST("/follow", ginx.WrapBodyAndToken(h.l, h.Follow))
	ug.POST("/cancel", ginx.WrapBodyAndToken(h.l, h.CanelFollow))
	ug.GET("/followee", ginx.WrapBodyAndToken(h.l, h.GetFollowee))
	ug.GET("/follower", ginx.WrapBodyAndToken(h.l, h.GetFollower))
	ug.GET("/static", ginx.WrapBodyAndToken(h.l, h.GetFollowStatic))
	ug.GET("/check", ginx.WrapBodyAndToken(h.l, h.CheckFollow)) // 检查关注状态
}

type FolloweeRequest struct {
	Followee int64 `json:"followee"`
}

type FollowerRequest struct {
	Follower int64 `json:"follower"`
}

// FollowUserVO 关注/粉丝列表中的用户信息（不含敏感信息）
type FollowUserVO struct {
	Id       int64  `json:"id"`
	Nickname string `json:"nickname"`
	AboutMe  string `json:"about_me"`
	Followed bool   `json:"followed"`
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

	// 发送关注通知（异步）
	go func() {
		// 使用新的 context，避免 HTTP 请求结束后 context 被取消
		newCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		// 获取关注者名称
		userResp, er := h.userSvc.Profile(newCtx, &userv1.ProfileRequest{Id: uc.UserId})
		followerName := "用户"
		if er == nil {
			followerName = userResp.GetUser().GetNickName()
		}

		er = h.notifyProducer.ProduceFollowEvent(newCtx, events.FollowEvent{
			FollowerId:   uc.UserId,
			FollowerName: followerName,
			FolloweeId:   req.Followee,
		})
		if er != nil {
			h.l.Error("发送关注通知失败", logger.Error(er))
		}
	}()

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
	follower := uc.UserId
	if req.Follower != 0 {
		follower = req.Follower
	}
	resp, err = h.svc.GetFollowee(ctx.Request.Context(), &followv1.GetFolloweeRequest{
		Follower: follower,
		Offset:   0,
		Limit:    1000,
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, nil
	}
	var res []FollowUserVO
	for _, relation := range resp.GetFollowRelations() {
		pResp, err := h.userSvc.Profile(ctx, &userv1.ProfileRequest{
			Id: relation.Followee,
		})
		if err != nil {
			continue
		}
		vo := FollowUserVO{
			Id:       pResp.GetUser().GetId(),
			Nickname: pResp.GetUser().GetNickName(),
			AboutMe:  pResp.GetUser().GetAboutMe(),
		}
		// 如果是查看自己的关注列表，所有人都是已关注
		if follower == uc.UserId {
			vo.Followed = true
		} else {
			// 查看他人的关注列表，检查当前用户是否关注了该用户
			infoResp, er := h.svc.FollowInfo(ctx.Request.Context(), &followv1.FollowInfoRequest{
				Follower: uc.UserId,
				Followee: relation.Followee,
			})
			if er == nil && infoResp.GetFollowRelation() != nil && infoResp.GetFollowRelation().GetId() > 0 {
				vo.Followed = true
			}
		}
		res = append(res, vo)
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
	followee := uc.UserId
	if req.Follower != 0 {
		followee = req.Follower
	}
	resp, err = h.svc.GetFollower(ctx.Request.Context(), &followv1.GetFollowerRequest{
		Followee: followee,
		Offset:   0,
		Limit:    1000,
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, nil
	}
	var res []FollowUserVO
	for _, relation := range resp.GetFollowRelations() {
		pResp, err := h.userSvc.Profile(ctx, &userv1.ProfileRequest{
			Id: relation.Follower,
		})
		if err != nil {
			continue
		}
		vo := FollowUserVO{
			Id:       pResp.GetUser().GetId(),
			Nickname: pResp.GetUser().GetNickName(),
			AboutMe:  pResp.GetUser().GetAboutMe(),
		}
		// 检查当前用户是否关注了该粉丝
		infoResp, er := h.svc.FollowInfo(ctx.Request.Context(), &followv1.FollowInfoRequest{
			Follower: uc.UserId,
			Followee: relation.Follower,
		})
		if er == nil && infoResp.GetFollowRelation() != nil && infoResp.GetFollowRelation().GetId() > 0 {
			vo.Followed = true
		}
		res = append(res, vo)
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

// CheckFollow 检查当前用户是否关注了指定用户
func (h *FollowHandler) CheckFollow(ctx *gin.Context, req FolloweeRequest, uc ijwt.UserClaims) (ginx.Result, error) {
	type Response struct {
		Followed bool `json:"followed"`
	}

	if req.Followee == 0 {
		return ginx.Result{Code: 4, Msg: "参数错误"}, nil
	}

	resp, err := h.svc.FollowInfo(ctx.Request.Context(), &followv1.FollowInfoRequest{
		Follower: uc.UserId,
		Followee: req.Followee,
	})

	if err != nil {
		// 可能是未关注，返回 false
		return ginx.Result{
			Code: 2,
			Msg:  "OK",
			Data: Response{Followed: false},
		}, nil
	}

	followed := resp.GetFollowRelation() != nil && resp.GetFollowRelation().GetId() > 0
	return ginx.Result{
		Code: 2,
		Msg:  "OK",
		Data: Response{Followed: followed},
	}, nil
}
