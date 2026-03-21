package domain

// MsgType 消息类型
type MsgType uint8

const (
	MsgTypeText  MsgType = 1 // 文本
	MsgTypeImage MsgType = 2 // 图片
)

func (t MsgType) String() string {
	switch t {
	case MsgTypeText:
		return "文本"
	case MsgTypeImage:
		return "图片"
	default:
		return "未知类型"
	}
}

// MsgStatus 消息状态
type MsgStatus uint8

const (
	MsgStatusSent      MsgStatus = 1 // 已发送
	MsgStatusDelivered MsgStatus = 2 // 已送达
	MsgStatusRead      MsgStatus = 3 // 已读
	MsgStatusRecalled  MsgStatus = 4 // 已撤回
)

func (s MsgStatus) String() string {
	switch s {
	case MsgStatusSent:
		return "已发送"
	case MsgStatusDelivered:
		return "已送达"
	case MsgStatusRead:
		return "已读"
	case MsgStatusRecalled:
		return "已撤回"
	default:
		return "未知状态"
	}
}
