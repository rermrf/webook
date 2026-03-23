package dao

import "context"

type CreditDAO interface {
	// 账户相关
	GetAccount(ctx context.Context, uid int64) (CreditAccount, error)
	CreateOrUpdateAccount(ctx context.Context, uid int64, changeAmt int64) (int64, error)

	// 流水相关
	CreateFlow(ctx context.Context, flow CreditFlow) error
	GetFlows(ctx context.Context, uid int64, offset, limit int) ([]CreditFlow, error)
	HasFlow(ctx context.Context, uid int64, biz string, bizId int64) (bool, error)

	// 每日限制相关
	GetDailyLimit(ctx context.Context, uid int64, biz string, date string) (DailyLimit, error)
	IncrDailyLimit(ctx context.Context, uid int64, biz string, date string, amt int64) error

	// 规则相关
	GetRules(ctx context.Context) ([]CreditRule, error)
	GetRule(ctx context.Context, biz string) (CreditRule, error)

	// 积分打赏相关
	CreateCreditReward(ctx context.Context, reward CreditReward) (int64, error)
	GetCreditReward(ctx context.Context, id int64) (CreditReward, error)
	UpdateCreditRewardStatus(ctx context.Context, id int64, status uint8) error

	// 转账（事务）
	TransferCredit(ctx context.Context, fromUid, toUid, amount int64, biz string, bizId int64) error
	// 全额转账（不抽成，开放API专用）
	TransferCreditFull(ctx context.Context, fromUid, toUid, amount int64, description string) error

	// 易支付订单相关
	CreateEpayOrder(ctx context.Context, order EpayOrder) (int64, error)
	GetEpayOrder(ctx context.Context, id int64) (EpayOrder, error)
	GetEpayOrderByTradeNo(ctx context.Context, tradeNo string) (EpayOrder, error)
	GetEpayOrderByOutTradeNo(ctx context.Context, appId, outTradeNo string) (EpayOrder, error)
	UpdateEpayOrderStatus(ctx context.Context, id int64, status uint8) error
	UpdateEpayOrderNotify(ctx context.Context, id int64, notifyCount int, notifyTime int64) error
	ListPendingNotifyOrders(ctx context.Context, limit int) ([]EpayOrder, error)
}

// CreditAccount 积分账户表
type CreditAccount struct {
	Id      int64 `gorm:"primaryKey,autoIncrement"`
	Uid     int64 `gorm:"uniqueIndex"`
	Balance int64
	Ctime   int64
	Utime   int64
}

func (CreditAccount) TableName() string {
	return "credit_accounts"
}

// CreditFlow 积分流水表
type CreditFlow struct {
	Id          int64  `gorm:"primaryKey,autoIncrement"`
	Uid         int64  `gorm:"index:idx_uid_ctime"`
	Biz         string `gorm:"uniqueIndex:idx_biz_bizid_uid;type:varchar(32)"`
	BizId       int64  `gorm:"uniqueIndex:idx_biz_bizid_uid"`
	ChangeAmt   int64
	Balance     int64
	Description string `gorm:"type:varchar(256)"`
	Ctime       int64  `gorm:"index:idx_uid_ctime"`
}

func (CreditFlow) TableName() string {
	return "credit_flows"
}

// DailyLimit 每日积分限制表
type DailyLimit struct {
	Id       int64  `gorm:"primaryKey,autoIncrement"`
	Uid      int64  `gorm:"uniqueIndex:idx_uid_biz_date"`
	Biz      string `gorm:"uniqueIndex:idx_uid_biz_date;type:varchar(32)"`
	Date     string `gorm:"uniqueIndex:idx_uid_biz_date;type:varchar(10)"`
	Count    int64
	TotalAmt int64
	Ctime    int64
	Utime    int64
}

func (DailyLimit) TableName() string {
	return "credit_daily_limits"
}

// CreditRule 积分规则配置表
type CreditRule struct {
	Id          int64  `gorm:"primaryKey,autoIncrement"`
	Biz         string `gorm:"uniqueIndex;type:varchar(32)"`
	CreditAmt   int64
	DailyLimit  int64
	Description string `gorm:"type:varchar(256)"`
	Enabled     bool   `gorm:"default:true"`
	Ctime       int64
	Utime       int64
}

func (CreditRule) TableName() string {
	return "credit_rules"
}

// CreditReward 积分打赏表
type CreditReward struct {
	Id        int64  `gorm:"primaryKey,autoIncrement"`
	Uid       int64  `gorm:"index"`
	TargetUid int64  `gorm:"index"`
	Biz       string `gorm:"type:varchar(32)"`
	BizId     int64
	Amount    int64
	Status    uint8
	Ctime     int64
	Utime     int64
}

func (CreditReward) TableName() string {
	return "credit_rewards"
}

// EpayOrder 易支付订单表
type EpayOrder struct {
	Id          int64  `gorm:"primaryKey,autoIncrement"`
	TradeNo     string `gorm:"uniqueIndex;type:varchar(64);not null;comment:平台订单号"`
	OutTradeNo  string `gorm:"index:idx_appid_outtradeno;type:varchar(64);not null;comment:商户订单号"`
	AppId       string `gorm:"index:idx_appid_outtradeno;type:varchar(64);not null;comment:商户ID"`
	Uid         int64  `gorm:"index;comment:支付用户ID"`
	Type        string `gorm:"type:varchar(32);default:credit;comment:支付类型"`
	Name        string `gorm:"type:varchar(256);comment:商品名称"`
	Money       int64  `gorm:"comment:金额(积分)"`
	Status      uint8  `gorm:"default:1;comment:状态1待支付2成功3关闭4已通知"`
	NotifyURL   string `gorm:"type:varchar(512);comment:异步通知地址"`
	ReturnURL   string `gorm:"type:varchar(512);comment:同步跳转地址"`
	Param       string `gorm:"type:varchar(512);comment:自定义参数"`
	NotifyCount int    `gorm:"default:0;comment:通知次数"`
	NotifyTime  int64  `gorm:"comment:最后通知时间"`
	Ctime       int64
	Utime       int64
}

func (EpayOrder) TableName() string {
	return "credit_epay_orders"
}
