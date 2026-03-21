package ws

// ClientMessage 客户端 → 服务端
type ClientMessage struct {
	Action         string `json:"action"`                    // send | ack | typing | heartbeat
	ConversationID string `json:"conversation_id,omitempty"`
	MsgType        uint8  `json:"msg_type,omitempty"`
	Content        string `json:"content,omitempty"`
	MessageID      string `json:"message_id,omitempty"`
	ReceiverId     int64  `json:"receiver_id,omitempty"`
}

// ServerMessage 服务端 → 客户端
type ServerMessage struct {
	Action         string       `json:"action"`                    // message | recall | ack | error
	ConversationID string       `json:"conversation_id,omitempty"`
	Message        *MessageData `json:"message,omitempty"`
	MessageID      string       `json:"message_id,omitempty"`
	Ctime          int64        `json:"ctime,omitempty"`
	Code           int          `json:"code,omitempty"`
	Msg            string       `json:"msg,omitempty"`
}

// MessageData 消息数据
type MessageData struct {
	Id         string `json:"id"`
	SenderId   int64  `json:"sender_id"`
	ReceiverId int64  `json:"receiver_id"`
	MsgType    uint8  `json:"msg_type"`
	Content    string `json:"content"`
	Ctime      int64  `json:"ctime"`
}
