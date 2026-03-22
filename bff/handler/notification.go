package handler

import (
	"io"
	"net/http"
	"strconv"
	"time"

	notificationv2 "webook/api/proto/gen/notification/v2"
	ijwt "webook/bff/handler/jwt"
	"webook/bff/handler/sse"
	"webook/pkg/ginx"
	"webook/pkg/logger"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	svc notificationv2.NotificationServiceClient
	hub *sse.Hub
	l   logger.LoggerV1
}

func NewNotificationHandler(
	svc notificationv2.NotificationServiceClient,
	hub *sse.Hub,
	l logger.LoggerV1,
) *NotificationHandler {
	return &NotificationHandler{
		svc: svc,
		hub: hub,
		l:   l,
	}
}

func (h *NotificationHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/notifications")

	// SSE 实时推送
	g.GET("/stream", h.Stream)

	// REST 接口
	g.GET("/list", ginx.WrapClaims(h.l, h.List))
	g.GET("/unread", ginx.WrapClaims(h.l, h.ListUnread))
	g.GET("/unread-count", ginx.WrapClaims(h.l, h.GetUnreadCount))
	g.POST("/read/:id", ginx.WrapClaims(h.l, h.MarkAsRead))
	g.POST("/read-all", ginx.WrapClaims(h.l, h.MarkAllAsRead))
	g.DELETE("/:id", ginx.WrapClaims(h.l, h.Delete))
}

// Stream SSE 实时推送通知
func (h *NotificationHandler) Stream(c *gin.Context) {
	// 获取用户信息
	uc, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, ginx.Result{Code: 4, Msg: "未登录"})
		return
	}
	claims := uc.(ijwt.UserClaims)

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("X-Accel-Buffering", "no") // 禁用 Nginx 缓冲

	// 创建客户端并注册
	client := sse.NewClient(claims.UserId)
	h.hub.Register(client)
	defer h.hub.Unregister(client)

	// 发送初始未读数
	h.sendUnreadCount(c, claims.UserId)

	// 心跳定时器
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// 监听客户端断开
	clientGone := c.Request.Context().Done()

	c.Stream(func(w io.Writer) bool {
		select {
		case <-clientGone:
			return false
		case <-ticker.C:
			// 发送心跳
			c.SSEvent("ping", "")
			return true
		case data := <-client.Channel:
			// 发送通知
			c.SSEvent("notification", string(data))
			return true
		}
	})
}

// sendUnreadCount 发送当前未读数
func (h *NotificationHandler) sendUnreadCount(c *gin.Context, userId int64) {
	resp, err := h.svc.GetUnreadCount(c.Request.Context(), &notificationv2.GetUnreadCountRequest{
		UserId: userId,
	})
	if err != nil {
		h.l.Error("获取未读数失败", logger.Error(err))
		return
	}

	byGroup := make(map[string]int64)
	for g, count := range resp.ByGroup {
		byGroup[notificationv2.NotificationGroup(g).String()] = count
	}

	msg := &sse.NotificationMessage{
		Type:    "unread_count",
		Total:   resp.Total,
		ByGroup: byGroup,
	}
	h.hub.SendToUser(userId, msg)
}

// List 获取通知列表
func (h *NotificationHandler) List(c *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	var req ListNotificationReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return ginx.Result{Code: 4, Msg: "参数错误"}, err
	}
	if req.Limit <= 0 || req.Limit > 50 {
		req.Limit = 20
	}

	resp, err := h.svc.ListNotifications(c.Request.Context(), &notificationv2.ListNotificationsRequest{
		UserId: uc.UserId,
		Offset: int32(req.Offset),
		Limit:  int32(req.Limit),
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
	}

	return ginx.Result{
		Code: 0,
		Data: h.toVOs(resp.Notifications),
	}, nil
}

// ListUnread 获取未读通知列表
func (h *NotificationHandler) ListUnread(c *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	var req ListNotificationReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return ginx.Result{Code: 4, Msg: "参数错误"}, err
	}
	if req.Limit <= 0 || req.Limit > 50 {
		req.Limit = 20
	}

	resp, err := h.svc.ListUnread(c.Request.Context(), &notificationv2.ListUnreadRequest{
		UserId: uc.UserId,
		Offset: int32(req.Offset),
		Limit:  int32(req.Limit),
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
	}

	return ginx.Result{
		Code: 0,
		Data: h.toVOs(resp.Notifications),
	}, nil
}

// GetUnreadCount 获取未读数
func (h *NotificationHandler) GetUnreadCount(c *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	resp, err := h.svc.GetUnreadCount(c.Request.Context(), &notificationv2.GetUnreadCountRequest{
		UserId: uc.UserId,
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
	}

	byGroup := make(map[string]int64)
	for g, count := range resp.ByGroup {
		byGroup[notificationv2.NotificationGroup(g).String()] = count
	}

	return ginx.Result{
		Code: 0,
		Data: UnreadCountVO{
			Total:   resp.Total,
			ByGroup: byGroup,
		},
	}, nil
}

// MarkAsRead 标记单条已读
func (h *NotificationHandler) MarkAsRead(c *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id == 0 {
		return ginx.Result{Code: 4, Msg: "参数错误"}, nil
	}

	_, err = h.svc.MarkAsRead(c.Request.Context(), &notificationv2.MarkAsReadRequest{
		UserId: uc.UserId,
		Id:     id,
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
	}

	return ginx.Result{Code: 0, Msg: "success"}, nil
}

// MarkAllAsRead 标记全部已读
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	_, err := h.svc.MarkAllAsRead(c.Request.Context(), &notificationv2.MarkAllAsReadRequest{
		UserId: uc.UserId,
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
	}

	return ginx.Result{Code: 0, Msg: "success"}, nil
}

// Delete 删除通知
func (h *NotificationHandler) Delete(c *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id == 0 {
		return ginx.Result{Code: 4, Msg: "参数错误"}, nil
	}

	_, err = h.svc.Delete(c.Request.Context(), &notificationv2.DeleteRequest{
		UserId: uc.UserId,
		Id:     id,
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
	}

	return ginx.Result{Code: 0, Msg: "success"}, nil
}

func (h *NotificationHandler) toVOs(notifications []*notificationv2.NotificationItem) []NotificationVO {
	vos := make([]NotificationVO, 0, len(notifications))
	for _, n := range notifications {
		vos = append(vos, NotificationVO{
			Id:          n.Id,
			GroupType:   n.GroupType.String(),
			SourceId:    n.SourceId,
			SourceName:  n.SourceName,
			TargetId:    n.TargetId,
			TargetType:  n.TargetType,
			TargetTitle: n.TargetTitle,
			Content:         n.Content,
			SourceAvatarUrl: n.GetSourceAvatarUrl(),
			IsRead:          n.IsRead,
			Ctime:       n.Ctime,
		})
	}
	return vos
}

// VO 和请求结构体
type ListNotificationReq struct {
	Offset int `form:"offset"`
	Limit  int `form:"limit"`
}

type NotificationVO struct {
	Id          int64  `json:"id"`
	GroupType   string `json:"group_type"`
	SourceId    int64  `json:"source_id"`
	SourceName  string `json:"source_name"`
	TargetId    int64  `json:"target_id"`
	TargetType  string `json:"target_type"`
	TargetTitle string `json:"target_title"`
	Content         string `json:"content"`
	SourceAvatarUrl string `json:"source_avatar_url,omitempty"`
	IsRead          bool   `json:"is_read"`
	Ctime       int64  `json:"ctime"`
}

type UnreadCountVO struct {
	Total   int64            `json:"total"`
	ByGroup map[string]int64 `json:"by_group"`
}
