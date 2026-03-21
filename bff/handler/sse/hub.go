package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
	"webook/pkg/logger"
)

// NotificationMessage SSE 推送的通知消息
type NotificationMessage struct {
	Type         string           `json:"type"`                      // 消息类型: notification/unread_count
	Total        int64            `json:"total"`                     // 未读总数
	ByGroup      map[string]int64 `json:"by_group,omitempty"`        // 按分组统计
	Notification *NotificationData `json:"notification,omitempty"`   // 具体通知
}

type NotificationData struct {
	Id          int64  `json:"id"`
	GroupType   int    `json:"group_type"`
	SourceId    int64  `json:"source_id"`
	SourceName  string `json:"source_name"`
	TargetId    int64  `json:"target_id"`
	TargetType  string `json:"target_type"`
	TargetTitle string `json:"target_title"`
	Content     string `json:"content"`
	Ctime       int64  `json:"ctime"`
}

// Client SSE 客户端连接
type Client struct {
	UserId  int64
	Channel chan []byte
}

// Hub SSE 连接管理器
type Hub struct {
	// clients 用户ID -> 客户端连接列表（一个用户可能有多个设备）
	clients map[int64]map[*Client]bool
	// register 注册通道
	register chan *Client
	// unregister 注销通道
	unregister chan *Client
	// broadcast 广播给指定用户
	broadcast chan *UserMessage
	// mu 保护 clients
	mu sync.RWMutex
	// redis 用于接收通知
	redis *redis.Client
	l     logger.LoggerV1
}

// UserMessage 发送给指定用户的消息
type UserMessage struct {
	UserId int64
	Data   []byte
}

func NewHub(redisClient redis.Cmdable, l logger.LoggerV1) *Hub {
	// 类型断言获取 *redis.Client，用于 Pub/Sub
	client, ok := redisClient.(*redis.Client)
	if !ok {
		panic("SSE Hub requires *redis.Client for Pub/Sub support")
	}
	return &Hub{
		clients:    make(map[int64]map[*Client]bool),
		register:   make(chan *Client, 256),
		unregister: make(chan *Client, 256),
		broadcast:  make(chan *UserMessage, 1024),
		redis:      client,
		l:          l,
	}
}

// Run 启动 Hub，处理注册、注销和广播
func (h *Hub) Run(ctx context.Context) {
	// 启动 Redis 订阅
	go h.subscribeRedis(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.UserId] == nil {
				h.clients[client.UserId] = make(map[*Client]bool)
			}
			h.clients[client.UserId][client] = true
			h.mu.Unlock()
			h.l.Debug("SSE client registered",
				logger.Int64("userId", client.UserId))

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.UserId]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Channel)
					if len(clients) == 0 {
						delete(h.clients, client.UserId)
					}
				}
			}
			h.mu.Unlock()
			h.l.Debug("SSE client unregistered",
				logger.Int64("userId", client.UserId))

		case msg := <-h.broadcast:
			h.mu.RLock()
			if clients, ok := h.clients[msg.UserId]; ok {
				for client := range clients {
					select {
					case client.Channel <- msg.Data:
					default:
						// 通道满了，跳过
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// subscribeRedis 订阅 Redis 通知频道
func (h *Hub) subscribeRedis(ctx context.Context) {
	pubsub := h.redis.Subscribe(ctx, "notification:sse")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			// 消息格式: {"user_id": int64, "data": <json object>}
			var envelope struct {
				UserId int64           `json:"user_id"`
				Data   json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal([]byte(msg.Payload), &envelope); err != nil {
				h.l.Error("解析 Redis 通知消息失败", logger.Error(err))
				continue
			}
			// 发送给指定用户
			h.broadcast <- &UserMessage{
				UserId: envelope.UserId,
				Data:   envelope.Data,
			}
		}
	}
}

// Register 注册客户端
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister 注销客户端
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// SendToUser 发送消息给指定用户
func (h *Hub) SendToUser(userId int64, msg *NotificationMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	h.broadcast <- &UserMessage{
		UserId: userId,
		Data:   data,
	}
	return nil
}

// GetOnlineCount 获取在线用户数
func (h *Hub) GetOnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// IsUserOnline 检查用户是否在线
func (h *Hub) IsUserOnline(userId int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	clients, ok := h.clients[userId]
	return ok && len(clients) > 0
}

// NewClient 创建新客户端
func NewClient(userId int64) *Client {
	return &Client{
		UserId:  userId,
		Channel: make(chan []byte, 64),
	}
}

// FormatSSE 格式化 SSE 消息
func FormatSSE(event string, data []byte) []byte {
	return []byte(fmt.Sprintf("event: %s\ndata: %s\n\n", event, data))
}
