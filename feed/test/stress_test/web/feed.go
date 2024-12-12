package web

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"webook/feed/domain"
	"webook/feed/service"
)

//1. 编写 http 接口连通系统

type FeedHandler struct {
	svc service.FeedService
}

func NewFeedHandler(svc service.FeedService) *FeedHandler {
	return &FeedHandler{svc: svc}
}

func (f *FeedHandler) RegisterRoutes(s *gin.Engine) {
	g := s.Group("/feed")
	g.POST("/list", f.FindFeedEventList)
	g.POST("/add", f.CreateFeedEvent)
}

func (f *FeedHandler) FindFeedEventList(ctx *gin.Context) {
	var req FindFeedEventReq
	err := ctx.ShouldBindQuery(&req)
	events, err := f.svc.GetFeedEventList(ctx, req.Uid, req.Timestamp, req.Limit)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "成功",
		Data: events,
	})
}

func (f *FeedHandler) CreateFeedEvent(ctx *gin.Context) {
	var req CreateFeedEventReq
	err := ctx.ShouldBindJSON(&req)
	var ext map[string]string
	err = json.Unmarshal([]byte(req.Ext), &ext)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
	err = f.svc.CreateFeedEvent(ctx, domain.FeedEvent{
		Type: req.Typ,
		Ext:  ext,
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
}
