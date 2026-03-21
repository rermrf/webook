package domain

import "fmt"

type Conversation struct {
	Id             string
	ConversationID string
	Members        []int64
	LastMsg        LastMessage
	Ctime          int64
	Utime          int64
}

type LastMessage struct {
	Content  string
	MsgType  MsgType
	SenderId int64
	Ctime    int64
}

// GenConversationID 生成会话ID: conv:{min}:{max}
func GenConversationID(userA, userB int64) string {
	if userA > userB {
		userA, userB = userB, userA
	}
	return fmt.Sprintf("conv:%d:%d", userA, userB)
}
