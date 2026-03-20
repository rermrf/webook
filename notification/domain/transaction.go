package domain

// Transaction 事务消息
type Transaction struct {
	Id                 int64
	NotificationId     int64
	Key                string
	BizId              string
	Status             TransactionStatus
	CheckBackTimeoutMs int64
	NextCheckTime      int64
	RetryCount         int
	MaxRetry           int
	Ctime              int64
	Utime              int64
}

// PrepareRequest 事务预提交请求
type PrepareRequest struct {
	Notification       Notification
	BizId              string
	CheckBackTimeoutMs int64
}
