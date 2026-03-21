package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	imv1 "webook/api/proto/gen/im/v1"
	"webook/pkg/logger"
)

// IMHub WebSocket 连接管理器 + Redis Pub/Sub 跨实例消息转发
type IMHub struct {
	clients    map[int64]map[*IMClient]bool
	register   chan *IMClient
	unregister chan *IMClient
	mu         sync.RWMutex
	redis      *redis.Client
	imSvc      imv1.IMServiceClient
	l          logger.LoggerV1
}

// NewIMHub 创建 IM Hub
func NewIMHub(redisClient redis.Cmdable, imSvc imv1.IMServiceClient, l logger.LoggerV1) *IMHub {
	client, ok := redisClient.(*redis.Client)
	if !ok {
		panic("IMHub requires *redis.Client for Pub/Sub")
	}
	return &IMHub{
		clients:    make(map[int64]map[*IMClient]bool),
		register:   make(chan *IMClient, 256),
		unregister: make(chan *IMClient, 256),
		redis:      client,
		imSvc:      imSvc,
		l:          l,
	}
}

// Run 启动 Hub，处理注册、注销和 Redis 订阅
func (h *IMHub) Run(ctx context.Context) {
	go h.subscribeRedis(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.UserId] == nil {
				h.clients[client.UserId] = make(map[*IMClient]bool)
			}
			h.clients[client.UserId][client] = true
			h.mu.Unlock()
			h.l.Debug("IM client registered",
				logger.Int64("userId", client.UserId))

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.UserId]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.clients, client.UserId)
					}
				}
			}
			h.mu.Unlock()
			h.l.Debug("IM client unregistered",
				logger.Int64("userId", client.UserId))
		}
	}
}

// HandleSend 处理客户端发送消息
func (h *IMHub) HandleSend(client *IMClient, msg *ClientMessage) {
	ctx := context.Background()
	resp, err := h.imSvc.SendMessage(ctx, &imv1.SendMessageRequest{
		SenderId:   client.UserId,
		ReceiverId: msg.ReceiverId,
		MsgType:    uint32(msg.MsgType),
		Content:    msg.Content,
	})
	if err != nil {
		h.sendToClient(client, &ServerMessage{
			Action: "error",
			Code:   5,
			Msg:    "发送失败: " + err.Error(),
		})
		return
	}

	// 发送 ack 给发送者
	h.sendToClient(client, &ServerMessage{
		Action:         "ack",
		MessageID:      resp.MessageId,
		ConversationID: resp.ConversationId,
		Ctime:          resp.Ctime,
	})

	// 构建推送消息给接收者
	pushMsg := &ServerMessage{
		Action:         "message",
		ConversationID: resp.ConversationId,
		Message: &MessageData{
			Id:         resp.MessageId,
			SenderId:   client.UserId,
			ReceiverId: msg.ReceiverId,
			MsgType:    msg.MsgType,
			Content:    msg.Content,
			Ctime:      resp.Ctime,
		},
	}
	pushData, err := json.Marshal(pushMsg)
	if err != nil {
		h.l.Error("序列化推送消息失败", logger.Error(err))
		return
	}

	// 通过 Redis Pub/Sub 发布消息
	envelope := struct {
		UserId int64           `json:"user_id"`
		Data   json.RawMessage `json:"data"`
	}{
		UserId: msg.ReceiverId,
		Data:   pushData,
	}
	payload, err := json.Marshal(envelope)
	if err != nil {
		h.l.Error("序列化 Redis 消息失败", logger.Error(err))
		return
	}
	h.redis.Publish(ctx, "im:msg:"+resp.ConversationId, payload)
}

// HandleHeartbeat 处理心跳
func (h *IMHub) HandleHeartbeat(client *IMClient) {
	ctx := context.Background()
	h.redis.Set(ctx, fmt.Sprintf("im:online:%d", client.UserId), "1", 30*time.Second)
}

// Register 注册客户端
func (h *IMHub) Register(client *IMClient) {
	h.register <- client
}

// Unregister 注销客户端
func (h *IMHub) Unregister(client *IMClient) {
	h.unregister <- client
}

// subscribeRedis 订阅 Redis 消息频道
func (h *IMHub) subscribeRedis(ctx context.Context) {
	pubsub := h.redis.PSubscribe(ctx, "im:msg:*")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			var envelope struct {
				UserId int64           `json:"user_id"`
				Data   json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal([]byte(msg.Payload), &envelope); err != nil {
				h.l.Error("解析 Redis IM 消息失败", logger.Error(err))
				continue
			}

			h.mu.RLock()
			if clients, ok := h.clients[envelope.UserId]; ok {
				for client := range clients {
					select {
					case client.Send <- envelope.Data:
					default:
						// 通道满了，跳过
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// sendToClient 发送消息给客户端
func (h *IMHub) sendToClient(client *IMClient, msg *ServerMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		h.l.Error("序列化消息失败", logger.Error(err))
		return
	}
	select {
	case client.Send <- data:
	default:
		// 通道满了，跳过
	}
}
