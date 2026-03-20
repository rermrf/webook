package domain

// Channel 渠道
type Channel uint8

const (
	ChannelInApp Channel = 1 // 站内通知
	ChannelSMS   Channel = 2 // 短信
	ChannelEmail Channel = 3 // 邮件
)

func (c Channel) String() string {
	switch c {
	case ChannelInApp:
		return "站内通知"
	case ChannelSMS:
		return "短信"
	case ChannelEmail:
		return "邮件"
	default:
		return "未知渠道"
	}
}

// NotificationStatus 通知状态
type NotificationStatus uint8

const (
	NotificationStatusInit    NotificationStatus = 1 // 初始化
	NotificationStatusSending NotificationStatus = 2 // 发送中
	NotificationStatusSent    NotificationStatus = 3 // 已发送
	NotificationStatusFailed  NotificationStatus = 4 // 发送失败
)

func (s NotificationStatus) String() string {
	switch s {
	case NotificationStatusInit:
		return "初始化"
	case NotificationStatusSending:
		return "发送中"
	case NotificationStatusSent:
		return "已发送"
	case NotificationStatusFailed:
		return "发送失败"
	default:
		return "未知状态"
	}
}

// TransactionStatus 事务状态
type TransactionStatus uint8

const (
	TransactionStatusPrepared  TransactionStatus = 1 // 已预提交
	TransactionStatusConfirmed TransactionStatus = 2 // 已确认
	TransactionStatusCancelled TransactionStatus = 3 // 已取消
)

func (s TransactionStatus) String() string {
	switch s {
	case TransactionStatusPrepared:
		return "已预提交"
	case TransactionStatusConfirmed:
		return "已确认"
	case TransactionStatusCancelled:
		return "已取消"
	default:
		return "未知事务状态"
	}
}

// SendStrategy 发送策略
type SendStrategy uint8

const (
	SendStrategyImmediate SendStrategy = 1 // 立即发送
	SendStrategyScheduled SendStrategy = 2 // 定时发送 (预留)
)

func (s SendStrategy) String() string {
	switch s {
	case SendStrategyImmediate:
		return "立即发送"
	case SendStrategyScheduled:
		return "定时发送"
	default:
		return "未知策略"
	}
}

// NotificationGroup 站内通知分组
type NotificationGroup uint8

const (
	NotificationGroupInteraction NotificationGroup = 1 // 互动消息
	NotificationGroupReply       NotificationGroup = 2 // 回复我的
	NotificationGroupMention     NotificationGroup = 3 // @我的
	NotificationGroupFollow      NotificationGroup = 4 // 关注
	NotificationGroupSystem      NotificationGroup = 5 // 系统通知
)

func (g NotificationGroup) String() string {
	switch g {
	case NotificationGroupInteraction:
		return "互动消息"
	case NotificationGroupReply:
		return "回复我的"
	case NotificationGroupMention:
		return "@我的"
	case NotificationGroupFollow:
		return "关注"
	case NotificationGroupSystem:
		return "系统通知"
	default:
		return "未知分组"
	}
}

// TemplateStatus 模板状态
type TemplateStatus uint8

const (
	TemplateStatusEnabled  TemplateStatus = 1 // 启用
	TemplateStatusDisabled TemplateStatus = 2 // 禁用
)

func (s TemplateStatus) String() string {
	switch s {
	case TemplateStatusEnabled:
		return "启用"
	case TemplateStatusDisabled:
		return "禁用"
	default:
		return "未知模板状态"
	}
}

// TransactionAction 事务回查结果
type TransactionAction uint8

const (
	TransactionActionCommit   TransactionAction = 1 // 确认发送
	TransactionActionRollback TransactionAction = 2 // 回滚取消
	TransactionActionPending  TransactionAction = 3 // 处理中
)

func (a TransactionAction) String() string {
	switch a {
	case TransactionActionCommit:
		return "确认发送"
	case TransactionActionRollback:
		return "回滚取消"
	case TransactionActionPending:
		return "处理中"
	default:
		return "未知事务动作"
	}
}
