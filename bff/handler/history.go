package handler

import (
	"github.com/gin-gonic/gin"

	historyv1 "webook/api/proto/gen/history/v1"
	ijwt "webook/bff/handler/jwt"
	"webook/pkg/ginx"
	"webook/pkg/logger"
)

type HistoryHandler struct {
	svc historyv1.HistoryServiceClient
	l   logger.LoggerV1
}

func NewHistoryHandler(svc historyv1.HistoryServiceClient, l logger.LoggerV1) *HistoryHandler {
	return &HistoryHandler{
		svc: svc,
		l:   l,
	}
}

func (h *HistoryHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/history")
	g.GET("/list", ginx.WrapClaims(h.l, h.List))
	g.DELETE("", ginx.WrapClaims(h.l, h.Clear))
}

// HistoryVO 浏览历史视图对象
type HistoryVO struct {
	Id         int64  `json:"id"`
	Biz        string `json:"biz"`
	BizId      int64  `json:"biz_id"`
	BizTitle   string `json:"biz_title"`
	AuthorName string `json:"author_name"`
	Ctime      int64  `json:"ctime"`
	Utime      int64  `json:"utime"`
}

// HistoryListReq 浏览历史分页请求
type HistoryListReq struct {
	Cursor int64 `form:"cursor"`
	Limit  int   `form:"limit"`
}

// List 获取浏览历史列表
func (h *HistoryHandler) List(c *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	var req HistoryListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return ginx.Result{Code: 4, Msg: "参数错误"}, err
	}
	if req.Limit <= 0 || req.Limit > 50 {
		req.Limit = 20
	}

	resp, err := h.svc.List(c.Request.Context(), &historyv1.ListRequest{
		UserId: uc.UserId,
		Cursor: req.Cursor,
		Limit:  int32(req.Limit),
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
	}

	vos := make([]HistoryVO, 0, len(resp.GetItems()))
	for _, item := range resp.GetItems() {
		vos = append(vos, HistoryVO{
			Id:         item.Id,
			Biz:        item.Biz,
			BizId:      item.BizId,
			BizTitle:   item.BizTitle,
			AuthorName: item.AuthorName,
			Ctime:      item.Ctime,
			Utime:      item.Utime,
		})
	}
	return ginx.Result{Data: map[string]any{
		"items":    vos,
		"has_more": resp.GetHasMore(),
	}}, nil
}

// Clear 清除浏览历史
func (h *HistoryHandler) Clear(c *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	_, err := h.svc.Clear(c.Request.Context(), &historyv1.ClearRequest{
		UserId: uc.UserId,
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
	}
	return ginx.Result{Msg: "OK"}, nil
}
