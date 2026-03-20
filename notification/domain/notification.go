package domain

// Notification 通知领域模型
type Notification struct {
	Id             int64
	Key            string
	BizId          string
	Channel        Channel
	Receiver       string
	UserId         int64
	TemplateId     string
	TemplateParams map[string]string
	Content        string
	Status         NotificationStatus
	Strategy       SendStrategy
	GroupType      NotificationGroup
	SourceId       int64
	SourceName     string
	TargetId       int64
	TargetType     string
	TargetTitle    string
	IsRead         bool
	Ctime          int64
	Utime          int64
}

// UnreadCount 未读数统计
type UnreadCount struct {
	Total   int64
	ByGroup map[NotificationGroup]int64
}
