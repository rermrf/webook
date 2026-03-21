package ws

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"

	"webook/pkg/logger"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 30 * time.Second
	maxMessageSize = 4096
)

// IMClient WebSocket 客户端连接
type IMClient struct {
	UserId int64
	Conn   *websocket.Conn
	Send   chan []byte
	hub    *IMHub
}

// NewIMClient 创建新的 IM 客户端
func NewIMClient(userId int64, conn *websocket.Conn, hub *IMHub) *IMClient {
	return &IMClient{
		UserId: userId,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		hub:    hub,
	}
}

// ReadPump 从 WebSocket 读取消息
func (c *IMClient) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				c.hub.l.Error("WebSocket 读取异常",
					logger.Int64("userId", c.UserId),
					logger.Error(err))
			}
			return
		}

		var msg ClientMessage
		if err = json.Unmarshal(message, &msg); err != nil {
			c.hub.l.Warn("WebSocket 消息解析失败",
				logger.Int64("userId", c.UserId),
				logger.Error(err))
			continue
		}

		switch msg.Action {
		case "send":
			c.hub.HandleSend(c, &msg)
		case "heartbeat":
			c.hub.HandleHeartbeat(c)
		case "ack":
			// no-op for now
		default:
			// ignore unknown actions
		}
	}
}

// WritePump 向 WebSocket 写入消息
func (c *IMClient) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// hub 关闭了 Send channel
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
