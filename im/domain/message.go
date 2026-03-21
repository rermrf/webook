package domain

type Message struct {
	Id             string
	ConversationID string
	SenderId       int64
	ReceiverId     int64
	MsgType        MsgType
	Content        string
	Status         MsgStatus
	Ctime          int64
}
