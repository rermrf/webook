package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	imv1 "webook/api/proto/gen/im/v1"
	userv1 "webook/api/proto/gen/user/v1"
	ijwt "webook/bff/handler/jwt"
	"webook/bff/handler/ws"
	"webook/pkg/ginx"
	"webook/pkg/logger"
)

type IMHandler struct {
	svc     imv1.IMServiceClient
	userSvc userv1.UserServiceClient
	hub     *ws.IMHub
	l       logger.LoggerV1
}

func NewIMHandler(svc imv1.IMServiceClient, userSvc userv1.UserServiceClient, hub *ws.IMHub, l logger.LoggerV1) *IMHandler {
	return &IMHandler{
		svc:     svc,
		userSvc: userSvc,
		hub:     hub,
		l:       l,
	}
}

func (h *IMHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/im")
	g.GET("/ws", h.WebSocket)
	g.GET("/conversations", ginx.WrapClaims(h.l, h.ListConversations))
	g.GET("/conversations/:id/messages", ginx.WrapClaims(h.l, h.ListMessages))
	g.POST("/conversations/:id/read", ginx.WrapClaims(h.l, h.MarkAsRead))
	g.GET("/unread-count", ginx.WrapClaims(h.l, h.GetUnreadCount))
	g.GET("/online/:userId", ginx.WrapClaims(h.l, h.GetOnlineStatus))
}

// WebSocket 处理 WebSocket 连接升级
func (h *IMHandler) WebSocket(c *gin.Context) {
	uc, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, ginx.Result{Code: 4, Msg: "未登录"})
		return
	}
	claims := uc.(ijwt.UserClaims)

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.l.Error("WebSocket upgrade failed", logger.Error(err))
		return
	}

	client := ws.NewIMClient(claims.UserId, conn, h.hub)
	h.hub.Register(client)
	go client.WritePump()
	go client.ReadPump()
}

// PeerUser 对端用户信息
type PeerUser struct {
	UserId   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// ConversationVO 会话视图对象
type ConversationVO struct {
	ConversationID string     `json:"conversation_id"`
	Members        []int64    `json:"members"`
	PeerUser       *PeerUser  `json:"peer_user,omitempty"`
	LastMsg        *MessageVO `json:"last_msg,omitempty"`
	UnreadCount    int64      `json:"unread_count"`
	Utime          int64      `json:"utime"`
}

// MessageVO 消息视图对象
type MessageVO struct {
	Id         string `json:"id"`
	SenderId   int64  `json:"sender_id"`
	ReceiverId int64  `json:"receiver_id"`
	MsgType    uint32 `json:"msg_type"`
	Content    string `json:"content"`
	Status     uint32 `json:"status"`
	Ctime      int64  `json:"ctime"`
}

// IMListReq IM 分页请求（基于游标）
type IMListReq struct {
	Cursor int64 `form:"cursor"`
	Limit  int   `form:"limit"`
}

// ListConversations 获取会话列表
func (h *IMHandler) ListConversations(c *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	var req IMListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return ginx.Result{Code: 4, Msg: "参数错误"}, err
	}
	if req.Limit <= 0 || req.Limit > 50 {
		req.Limit = 20
	}

	resp, err := h.svc.ListConversations(c.Request.Context(), &imv1.ListConversationsRequest{
		UserId: uc.UserId,
		Cursor: req.Cursor,
		Limit:  int32(req.Limit),
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
	}

	// Collect peer user IDs
	peerIds := make([]int64, 0)
	for _, conv := range resp.Conversations {
		for _, memberId := range conv.Members {
			if memberId != uc.UserId {
				peerIds = append(peerIds, memberId)
			}
		}
	}

	// Batch fetch user profiles
	peerMap := make(map[int64]*PeerUser)
	for _, peerId := range peerIds {
		profile, profileErr := h.userSvc.Profile(c.Request.Context(), &userv1.ProfileRequest{Id: peerId})
		if profileErr == nil && profile.GetUser() != nil {
			peerMap[peerId] = &PeerUser{
				UserId:   peerId,
				Nickname: profile.GetUser().GetNickName(),
			}
		}
	}

	// Build VOs with PeerUser
	vos := make([]ConversationVO, 0, len(resp.Conversations))
	for _, conv := range resp.Conversations {
		vo := ConversationVO{
			ConversationID: conv.ConversationId,
			Members:        conv.Members,
			UnreadCount:    conv.UnreadCount,
			Utime:          conv.Utime,
		}
		for _, memberId := range conv.Members {
			if memberId != uc.UserId {
				vo.PeerUser = peerMap[memberId]
				break
			}
		}
		if conv.LastMsg != nil {
			vo.LastMsg = h.toMessageVO(conv.LastMsg)
		}
		vos = append(vos, vo)
	}
	return ginx.Result{Data: vos}, nil
}

// ListMessages 获取历史消息
func (h *IMHandler) ListMessages(c *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	convId := c.Param("id")
	if convId == "" {
		return ginx.Result{Code: 4, Msg: "参数错误"}, nil
	}

	var req IMListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return ginx.Result{Code: 4, Msg: "参数错误"}, err
	}
	if req.Limit <= 0 || req.Limit > 50 {
		req.Limit = 20
	}

	resp, err := h.svc.ListMessages(c.Request.Context(), &imv1.ListMessagesRequest{
		ConversationId: convId,
		Cursor:         req.Cursor,
		Limit:          int32(req.Limit),
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
	}

	vos := make([]MessageVO, 0, len(resp.Messages))
	for _, msg := range resp.Messages {
		vos = append(vos, *h.toMessageVO(msg))
	}
	return ginx.Result{Data: vos}, nil
}

// MarkAsRead 标记会话已读
func (h *IMHandler) MarkAsRead(c *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	convId := c.Param("id")
	if convId == "" {
		return ginx.Result{Code: 4, Msg: "参数错误"}, nil
	}

	_, err := h.svc.MarkAsRead(c.Request.Context(), &imv1.MarkAsReadRequest{
		UserId:         uc.UserId,
		ConversationId: convId,
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
	}
	return ginx.Result{Msg: "OK"}, nil
}

// GetUnreadCount 获取未读消息数
func (h *IMHandler) GetUnreadCount(c *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	resp, err := h.svc.GetUnreadCount(c.Request.Context(), &imv1.GetUnreadCountRequest{
		UserId: uc.UserId,
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
	}
	return ginx.Result{Data: map[string]any{
		"total":           resp.Total,
		"by_conversation": resp.ByConversation,
	}}, nil
}

func (h *IMHandler) toMessageVO(msg *imv1.MessageItem) *MessageVO {
	return &MessageVO{
		Id:         msg.Id,
		SenderId:   msg.SenderId,
		ReceiverId: msg.ReceiverId,
		MsgType:    msg.MsgType,
		Content:    msg.Content,
		Status:     msg.Status,
		Ctime:      msg.Ctime,
	}
}

// GetOnlineStatus 查询用户在线状态
func (h *IMHandler) GetOnlineStatus(c *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	userIdStr := c.Param("userId")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		return ginx.Result{Code: 4, Msg: "参数错误"}, nil
	}
	resp, err := h.svc.IsOnline(c.Request.Context(), &imv1.IsOnlineRequest{UserId: userId})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
	}
	return ginx.Result{Data: map[string]bool{"online": resp.Online}}, nil
}
