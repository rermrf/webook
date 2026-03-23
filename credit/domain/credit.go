package domain

// CreditAccount 积分账户
type CreditAccount struct {
	Id      int64
	Uid     int64
	Balance int64 // 当前积分余额
	Ctime   int64
	Utime   int64
}

// CreditFlow 积分流水记录
type CreditFlow struct {
	Id          int64
	Uid         int64
	Biz         string // 业务类型: read/like/collect/comment/recharge/reward_in/reward_out
	BizId       int64  // 业务ID
	ChangeAmt   int64  // 变动金额，正为收入负为支出
	Balance     int64  // 变动后余额
	Description string // 描述
	Ctime       int64
}

// DailyLimit 每日积分限制
type DailyLimit struct {
	Id       int64
	Uid      int64
	Biz      string // 业务类型
	Date     string // YYYY-MM-DD
	Count    int64  // 今日获取次数
	TotalAmt int64  // 今日获取总积分
	Ctime    int64
	Utime    int64
}

// CreditRule 积分规则配置
type CreditRule struct {
	Id          int64
	Biz         string // 业务类型
	CreditAmt   int64  // 每次获取积分数
	DailyLimit  int64  // 每日上限，0表示无限制
	Description string
	Enabled     bool
	Ctime       int64
	Utime       int64
}

// CreditReward 积分打赏记录
type CreditReward struct {
	Id        int64
	Uid       int64              // 打赏者
	TargetUid int64              // 被打赏者
	Biz       string             // 业务类型
	BizId     int64              // 业务ID
	Amount    int64              // 打赏积分数量
	Status    CreditRewardStatus // 打赏状态
	Ctime     int64
	Utime     int64
}

// CreditRewardStatus 积分打赏状态
type CreditRewardStatus uint8

func (s CreditRewardStatus) AsUint8() uint8 {
	return uint8(s)
}

const (
	CreditRewardStatusUnknown CreditRewardStatus = iota
	CreditRewardStatusPending                    // 处理中
	CreditRewardStatusSuccess                    // 成功
	CreditRewardStatusFailed                     // 失败
)

// DailyStatus 每日积分状态（用于返回）
type DailyStatus struct {
	Biz         string
	EarnedCount int64 // 今日已获取次数
	EarnedAmt   int64 // 今日已获取积分
	DailyLimit  int64 // 每日上限
	Remaining   int64 // 剩余可获取积分
}

// EpayOrder 易支付订单
type EpayOrder struct {
	Id          int64
	TradeNo     string          // 平台订单号
	OutTradeNo  string          // 商户订单号
	AppId       string          // 商户ID (对应openapi的app_id)
	Uid         int64           // 支付用户ID
	Type        string          // 支付类型
	Name        string          // 商品名称
	Money       int64           // 金额（积分数量）
	Status      EpayOrderStatus // 订单状态
	NotifyURL   string          // 异步通知地址
	ReturnURL   string          // 同步跳转地址
	Param       string          // 自定义参数
	NotifyCount int             // 通知次数
	NotifyTime  int64           // 最后通知时间
	Ctime       int64
	Utime       int64
}

// EpayOrderStatus 易支付订单状态
type EpayOrderStatus uint8

const (
	EpayOrderStatusUnknown EpayOrderStatus = iota
	EpayOrderStatusWait                    // 待支付
	EpayOrderStatusSuccess                 // 支付成功
	EpayOrderStatusClosed                  // 已关闭
	EpayOrderStatusNotified                // 已通知成功
)
