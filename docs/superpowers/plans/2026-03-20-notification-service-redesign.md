# 通知服务重构实现计划

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 重构通知微服务为统一通知网关，支持站内/SMS/Email渠道、模板管理、TCC分布式事务和事务回查机制。

**Architecture:** 渠道策略模式——NotificationService 作为编排层，通过 ChannelSender 接口路由到 InApp/SMS/Email 实现。TCC 事务通过独立的 notification_transactions 表管理状态，CheckBackScheduler 定时回查超时记录。模板通过 TemplateService 管理 CRUD 和渲染。

**Tech Stack:** Go, gRPC/protobuf, GORM, Redis, Kafka (Sarama), Wire DI, ETCD, text/template

**Spec:** `docs/superpowers/specs/2026-03-20-notification-service-redesign.md`

---

## File Map

### 新建文件

| 文件路径 | 职责 |
|---------|------|
| `notification/domain/types.go` | 所有枚举定义（Channel, NotificationStatus, TransactionStatus, SendStrategy, NotificationGroup, TemplateStatus, TransactionAction） |
| `notification/domain/template.go` | Template 领域模型 |
| `notification/domain/transaction.go` | Transaction、PrepareRequest 领域模型 |
| `notification/service/template.go` | TemplateService 接口与实现（CRUD + 渲染） |
| `notification/repository/template.go` | CachedTemplateRepository |
| `notification/repository/transaction.go` | TransactionRepository |
| `notification/repository/dao/template.go` | Template DAO（GORM） |
| `notification/repository/dao/transaction.go` | Transaction DAO（GORM） |
| `notification/repository/cache/template.go` | Template Redis 缓存 |
| `notification/channel/types.go` | ChannelSender 接口定义 |
| `notification/channel/inapp.go` | InAppSender 实现 |
| `notification/channel/sms.go` | SMSSender 实现 |
| `notification/channel/sms_provider.go` | SMSProvider 接口 + 阿里云/腾讯云实现 |
| `notification/channel/email.go` | EmailSender 预留实现 |
| `notification/scheduler/checkback.go` | CheckBackScheduler 事务回查定时任务 |
| `notification/ioc/channel.go` | 渠道 Sender 初始化 |
| `notification/ioc/scheduler.go` | 回查任务初始化 |
| `notification/ioc/template.go` | 模板服务初始化 |
| `api/proto/notification/v2/notification.proto` | 新版 proto 定义（v2） |
| `api/proto/notification/v2/transaction_checker.proto` | TransactionChecker proto |

### 修改文件

| 文件路径 | 改动说明 |
|---------|---------|
| `notification/domain/notification.go` | 重写 Notification 领域模型，替换 NotificationType 为 NotificationGroup，新增 Key/Channel/TemplateId 等字段 |
| `notification/service/notification.go` | 重写 NotificationService，新增 Send 路由、Prepare/Confirm/Cancel TCC 接口 |
| `notification/repository/notification.go` | 重写 Repository 接口和 CachedNotificationRepository，适配新模型 |
| `notification/repository/dao/notification.go` | 重写 DAO entity 和方法，适配新表结构 |
| `notification/repository/dao/init.go` | 新增 Template、Transaction 表的 AutoMigrate |
| `notification/repository/cache/notification.go` | 改为按 NotificationGroup 维度统计未读计数 |
| `notification/grpc/server.go` | 重写 gRPC server，适配新 proto v2 |
| `notification/events/types.go` | 保留现有事件类型不变 |
| `notification/events/consumer.go` | 消费者改为调用 NotificationService.Send |
| `notification/wire.go` | 新增 templateSet、channelSet、schedulerSet |
| `notification/wire_gen.go` | Wire 重新生成 |
| `notification/app.go` | 新增 scheduler 字段 |
| `notification/main.go` | 启动 scheduler |
| `notification/ioc/kafka.go` | 消费者构造函数参数变更 |
| `script/mysql/init.sql` | 新增三张表的建表语句 |

---

## Chunk 1: Domain 层 + 枚举 + Proto 定义

### Task 1: 创建枚举定义 types.go

**Files:**
- Create: `notification/domain/types.go`

- [ ] **Step 1: 创建枚举定义文件**

```go
// notification/domain/types.go
package domain

// ===== 渠道 =====
type Channel uint8

const (
	ChannelInApp Channel = 1
	ChannelSMS   Channel = 2
	ChannelEmail Channel = 3
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

// ===== 通知状态 =====
type NotificationStatus uint8

const (
	NotificationStatusInit    NotificationStatus = 1
	NotificationStatusSending NotificationStatus = 2
	NotificationStatusSent    NotificationStatus = 3
	NotificationStatusFailed  NotificationStatus = 4
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

// ===== 事务状态 =====
type TransactionStatus uint8

const (
	TransactionStatusPrepared  TransactionStatus = 1
	TransactionStatusConfirmed TransactionStatus = 2
	TransactionStatusCancelled TransactionStatus = 3
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
		return "未知状态"
	}
}

// ===== 发送策略 =====
type SendStrategy uint8

const (
	SendStrategyImmediate SendStrategy = 1
	SendStrategyScheduled SendStrategy = 2 // 预留
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

// ===== 站内通知分组 =====
type NotificationGroup uint8

const (
	NotificationGroupInteraction NotificationGroup = 1
	NotificationGroupReply       NotificationGroup = 2
	NotificationGroupMention     NotificationGroup = 3
	NotificationGroupFollow      NotificationGroup = 4
	NotificationGroupSystem      NotificationGroup = 5
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

// ===== 模板状态 =====
type TemplateStatus uint8

const (
	TemplateStatusEnabled  TemplateStatus = 1
	TemplateStatusDisabled TemplateStatus = 2
)

func (s TemplateStatus) String() string {
	switch s {
	case TemplateStatusEnabled:
		return "启用"
	case TemplateStatusDisabled:
		return "禁用"
	default:
		return "未知状态"
	}
}

// ===== 事务回查结果 =====
type TransactionAction uint8

const (
	TransactionActionCommit   TransactionAction = 1
	TransactionActionRollback TransactionAction = 2
	TransactionActionPending  TransactionAction = 3
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
		return "未知操作"
	}
}
```

- [ ] **Step 2: 验证编译**

Run: `cd notification && go build ./domain/...`
Expected: 编译通过，无错误

- [ ] **Step 3: 提交**

```bash
git add notification/domain/types.go
git commit -m "feat(notification): 添加枚举定义 types.go"
```

---

### Task 2: 重写 Notification 领域模型

**Files:**
- Modify: `notification/domain/notification.go`

- [ ] **Step 1: 重写 notification.go**

删除旧的 `NotificationType` 枚举和 `UnreadCount` 结构体，替换为新的 `Notification` 模型。

```go
// notification/domain/notification.go
package domain

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
	// 站内通知专用
	GroupType   NotificationGroup
	SourceId    int64
	SourceName  string
	TargetId    int64
	TargetType  string
	TargetTitle string
	IsRead      bool
	Ctime       int64
	Utime       int64
}

type UnreadCount struct {
	Total   int64
	ByGroup map[NotificationGroup]int64
}
```

- [ ] **Step 2: 验证编译**（此时会有编译错误因为下游代码还引用旧类型，属于预期行为）

Run: `cd notification && go build ./domain/...`
Expected: domain 包编译通过

- [ ] **Step 3: 提交**

```bash
git add notification/domain/notification.go
git commit -m "feat(notification): 重写 Notification 领域模型"
```

---

### Task 3: 创建 Template 和 Transaction 领域模型

**Files:**
- Create: `notification/domain/template.go`
- Create: `notification/domain/transaction.go`

- [ ] **Step 1: 创建 template.go**

```go
// notification/domain/template.go
package domain

type Template struct {
	Id                    int64
	TemplateId            string
	Channel               Channel
	Name                  string
	Content               string
	Description           string
	Status                TemplateStatus
	SMSSign               string
	SMSProviderTemplateId string
	Ctime                 int64
	Utime                 int64
}
```

- [ ] **Step 2: 创建 transaction.go**

```go
// notification/domain/transaction.go
package domain

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

type PrepareRequest struct {
	Notification       Notification
	BizId              string
	CheckBackTimeoutMs int64
}
```

- [ ] **Step 3: 验证编译**

Run: `cd notification && go build ./domain/...`
Expected: 编译通过

- [ ] **Step 4: 提交**

```bash
git add notification/domain/template.go notification/domain/transaction.go
git commit -m "feat(notification): 添加 Template 和 Transaction 领域模型"
```

---

### Task 4: 编写新版 Proto 定义

**Files:**
- Create: `api/proto/notification/v2/notification.proto`
- Create: `api/proto/notification/v2/transaction_checker.proto`

- [ ] **Step 1: 创建 notification.proto v2**

完整内容按 spec 中"gRPC API 设计"章节的所有枚举、消息、服务定义编写。包含：
- 所有枚举（Channel, SendStrategy, NotificationStatus, TransactionStatus, TransactionAction, NotificationGroup）
- Notification 消息体
- NotificationService 服务（18个 RPC 方法）
- 所有 Request/Response 消息
- Template 消息、NotificationItem 消息

Proto package: `api.notification.v2`
Go package: `gen/notification/v2;notificationv2`

- [ ] **Step 2: 创建 transaction_checker.proto**

```protobuf
syntax = "proto3";
package api.notification.v2;
option go_package = "gen/notification/v2;notificationv2";

import "notification/v2/notification.proto";

service TransactionChecker {
  rpc CheckTransaction(CheckTransactionRequest) returns (CheckTransactionResponse);
}

message CheckTransactionRequest {
  string key = 1;
}

message CheckTransactionResponse {
  TransactionAction action = 1;
}
```

- [ ] **Step 3: 生成 Go 代码**

Run: `make proto` 或项目中对应的 protobuf 生成命令（参考 Makefile）
Expected: `api/proto/gen/notification/v2/` 下生成 `notification.pb.go`、`notification_grpc.pb.go`、`transaction_checker.pb.go`、`transaction_checker_grpc.pb.go`

- [ ] **Step 4: 验证生成代码编译**

Run: `go build ./api/proto/gen/notification/v2/...`
Expected: 编译通过

- [ ] **Step 5: 提交**

```bash
git add api/proto/notification/v2/ api/proto/gen/notification/v2/
git commit -m "feat(notification): 添加 v2 proto 定义和生成代码"
```

---

## Chunk 2: DAO 层 + Repository 层

### Task 5: 重写 Notification DAO

**Files:**
- Modify: `notification/repository/dao/notification.go`
- Modify: `notification/repository/dao/init.go`

- [ ] **Step 1: 重写 DAO entity 和接口**

```go
// notification/repository/dao/notification.go
package dao

import (
	"context"
	"encoding/json"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Notification struct {
	Id             int64           `gorm:"primaryKey;autoIncrement"`
	KeyField       string          `gorm:"column:key_field;type:varchar(256);uniqueIndex:uk_key_channel"`
	BizId          string          `gorm:"type:varchar(64)"`
	Channel        uint8           `gorm:"uniqueIndex:uk_key_channel"`
	Receiver       string          `gorm:"type:varchar(256);index:idx_receiver_channel"`
	UserId         int64           `gorm:"index:idx_user_ctime;index:idx_user_group;index:idx_user_unread"`
	TemplateId     string          `gorm:"type:varchar(128)"`
	TemplateParams json.RawMessage `gorm:"type:json"`
	Content        string          `gorm:"type:text"`
	Status         uint8
	Strategy       uint8
	GroupType      uint8  `gorm:"index:idx_user_group"`
	SourceId       int64
	SourceName     string `gorm:"type:varchar(128)"`
	TargetId       int64
	TargetType     string `gorm:"type:varchar(64)"`
	TargetTitle    string `gorm:"type:varchar(256)"`
	IsRead         bool   `gorm:"index:idx_user_unread;default:false"`
	Ctime          int64  `gorm:"index:idx_user_ctime;index:idx_receiver_channel"`
	Utime          int64
}

type NotificationDAO interface {
	Insert(ctx context.Context, n Notification) (int64, error)
	BatchInsert(ctx context.Context, ns []Notification) ([]int64, error)
	FindByKeyAndChannel(ctx context.Context, key string, channel uint8) (Notification, error)
	FindByUserId(ctx context.Context, userId int64, offset, limit int) ([]Notification, error)
	FindByUserIdAndGroup(ctx context.Context, userId int64, groupType uint8, offset, limit int) ([]Notification, error)
	FindUnreadByUserId(ctx context.Context, userId int64, offset, limit int) ([]Notification, error)
	MarkAsRead(ctx context.Context, userId int64, ids []int64) error
	MarkAllAsRead(ctx context.Context, userId int64) error
	CountUnreadByGroup(ctx context.Context, userId int64) (map[uint8]int64, error)
	UpdateStatus(ctx context.Context, id int64, status uint8) error
	Delete(ctx context.Context, userId int64, id int64) error
	DeleteByUserId(ctx context.Context, userId int64) error
}

type GORMNotificationDAO struct {
	db *gorm.DB
}

func NewGORMNotificationDAO(db *gorm.DB) NotificationDAO {
	return &GORMNotificationDAO{db: db}
}
```

实现所有方法：
- `Insert`：插入单条，使用 `clause.OnConflict{DoNothing: true}` 做幂等
- `BatchInsert`：批量插入
- `FindByKeyAndChannel`：按幂等键查询
- `FindByUserId`：按 user_id + ctime DESC 分页
- `FindByUserIdAndGroup`：按 user_id + group_type + ctime DESC 分页
- `FindUnreadByUserId`：is_read = false
- `MarkAsRead`：按 ids 批量更新 is_read
- `MarkAllAsRead`：按 user_id 全量更新
- `CountUnreadByGroup`：`GROUP BY group_type` 统计
- `UpdateStatus`：更新 status 字段
- `Delete` / `DeleteByUserId`

- [ ] **Step 2: 验证编译**

Run: `cd notification && go build ./repository/dao/...`
Expected: 编译通过

- [ ] **Step 3: 提交**

```bash
git add notification/repository/dao/notification.go
git commit -m "feat(notification): 重写 Notification DAO 适配新表结构"
```

---

### Task 6: 创建 Template DAO

**Files:**
- Create: `notification/repository/dao/template.go`

- [ ] **Step 1: 创建 Template DAO**

```go
// notification/repository/dao/template.go
package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type NotificationTemplate struct {
	Id                    int64  `gorm:"primaryKey;autoIncrement"`
	TemplateId            string `gorm:"type:varchar(128);uniqueIndex:uk_template_channel"`
	Channel               uint8  `gorm:"uniqueIndex:uk_template_channel;index:idx_channel_status"`
	Name                  string `gorm:"type:varchar(256)"`
	Content               string `gorm:"type:text;not null"`
	Description           string `gorm:"type:varchar(512)"`
	Status                uint8  `gorm:"index:idx_channel_status;default:1"`
	SMSSign               string `gorm:"column:sms_sign;type:varchar(64)"`
	SMSProviderTemplateId string `gorm:"column:sms_provider_template_id;type:varchar(128)"`
	Ctime                 int64
	Utime                 int64
}

type TemplateDAO interface {
	Insert(ctx context.Context, t NotificationTemplate) (int64, error)
	Update(ctx context.Context, t NotificationTemplate) error
	FindByTemplateIdAndChannel(ctx context.Context, templateId string, channel uint8) (NotificationTemplate, error)
	FindByChannel(ctx context.Context, channel uint8, offset, limit int) ([]NotificationTemplate, error)
}

type GORMTemplateDAO struct {
	db *gorm.DB
}

func NewGORMTemplateDAO(db *gorm.DB) TemplateDAO {
	return &GORMTemplateDAO{db: db}
}
```

实现所有方法：
- `Insert`：设置 Ctime/Utime，插入
- `Update`：按 template_id + channel 更新，设置 Utime
- `FindByTemplateIdAndChannel`：唯一查询
- `FindByChannel`：按 channel 分页查询，只返回 status=1(启用)

- [ ] **Step 2: 验证编译**

Run: `cd notification && go build ./repository/dao/...`
Expected: 编译通过

- [ ] **Step 3: 提交**

```bash
git add notification/repository/dao/template.go
git commit -m "feat(notification): 添加 Template DAO"
```

---

### Task 7: 创建 Transaction DAO

**Files:**
- Create: `notification/repository/dao/transaction.go`

- [ ] **Step 1: 创建 Transaction DAO**

```go
// notification/repository/dao/transaction.go
package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type NotificationTransaction struct {
	Id                 int64  `gorm:"primaryKey;autoIncrement"`
	NotificationId     int64  `gorm:"uniqueIndex:uk_notification_id"`
	KeyField           string `gorm:"column:key_field;type:varchar(256);uniqueIndex:uk_key"`
	BizId              string `gorm:"type:varchar(64)"`
	Status             uint8  `gorm:"index:idx_status_check"`
	CheckBackTimeoutMs int64
	NextCheckTime      int64 `gorm:"index:idx_status_check"`
	RetryCount         int
	MaxRetry           int `gorm:"default:5"`
	Ctime              int64
	Utime              int64
}

type TransactionDAO interface {
	Insert(ctx context.Context, t NotificationTransaction) (int64, error)
	FindByKey(ctx context.Context, key string) (NotificationTransaction, error)
	UpdateStatus(ctx context.Context, key string, status uint8) error
	FindPreparedTimeout(ctx context.Context, now int64, limit int) ([]NotificationTransaction, error)
	IncrRetryCount(ctx context.Context, id int64, nextCheckTime int64) error
}

type GORMTransactionDAO struct {
	db *gorm.DB
}

func NewGORMTransactionDAO(db *gorm.DB) TransactionDAO {
	return &GORMTransactionDAO{db: db}
}
```

实现所有方法：
- `Insert`：插入事务记录
- `FindByKey`：按 key_field 查询
- `UpdateStatus`：更新 status + utime
- `FindPreparedTimeout`：`status=1 AND next_check_time < now`，按 next_check_time ASC，限制 limit 条
- `IncrRetryCount`：`retry_count = retry_count + 1`，更新 next_check_time

- [ ] **Step 2: 更新 init.go 添加新表迁移**

```go
// notification/repository/dao/init.go
package dao

import "gorm.io/gorm"

func InitTable(db *gorm.DB) error {
	return db.AutoMigrate(
		&Notification{},
		&NotificationTemplate{},
		&NotificationTransaction{},
	)
}
```

- [ ] **Step 3: 验证编译**

Run: `cd notification && go build ./repository/dao/...`
Expected: 编译通过

- [ ] **Step 4: 提交**

```bash
git add notification/repository/dao/transaction.go notification/repository/dao/init.go
git commit -m "feat(notification): 添加 Transaction DAO 并更新表迁移"
```

---

### Task 8: 重写 Notification Cache

**Files:**
- Modify: `notification/repository/cache/notification.go`

- [ ] **Step 1: 重写 Cache 接口和实现**

改为按 `NotificationGroup` 维度统计未读计数。Redis Hash 的 field 从旧的 `{typeNum}` 改为 `{groupNum}`。

接口保持：
- `IncrUnreadCount(ctx, userId int64, group uint8) error`
- `GetUnreadCount(ctx, userId int64) (map[uint8]int64, int64, error)` — 返回 byGroup map 和 total
- `ClearUnreadCount(ctx, userId int64) error`
- `SetUnreadCount(ctx, userId int64, counts map[uint8]int64) error`
- `PublishSSE(ctx, userId int64, data []byte) error` — 发布原始 JSON 到 SSE 通道

Redis key: `notification:unread:{userId}`，Hash 结构不变，field 改为 group 编号。

- [ ] **Step 2: 验证编译**

Run: `cd notification && go build ./repository/cache/...`
Expected: 编译通过

- [ ] **Step 3: 提交**

```bash
git add notification/repository/cache/notification.go
git commit -m "feat(notification): 重写 Cache 按 NotificationGroup 维度统计未读"
```

---

### Task 9: 创建 Template Cache

**Files:**
- Create: `notification/repository/cache/template.go`

- [ ] **Step 1: 创建 Template Cache**

```go
// notification/repository/cache/template.go
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type TemplateCache interface {
	Get(ctx context.Context, templateId string, channel uint8) ([]byte, error)
	Set(ctx context.Context, templateId string, channel uint8, data []byte) error
	Del(ctx context.Context, templateId string, channel uint8) error
}

type RedisTemplateCache struct {
	client redis.Cmdable
}

func NewRedisTemplateCache(client redis.Cmdable) TemplateCache {
	return &RedisTemplateCache{client: client}
}

func (c *RedisTemplateCache) key(templateId string, channel uint8) string {
	return fmt.Sprintf("notification:template:%s:%d", templateId, channel)
}

func (c *RedisTemplateCache) Get(ctx context.Context, templateId string, channel uint8) ([]byte, error) {
	data, err := c.client.Get(ctx, c.key(templateId, channel)).Bytes()
	if err == redis.Nil {
		return nil, ErrKeyNotExist
	}
	return data, err
}

func (c *RedisTemplateCache) Set(ctx context.Context, templateId string, channel uint8, data []byte) error {
	return c.client.Set(ctx, c.key(templateId, channel), data, 24*time.Hour).Err()
}

func (c *RedisTemplateCache) Del(ctx context.Context, templateId string, channel uint8) error {
	return c.client.Del(ctx, c.key(templateId, channel)).Err()
}
```

- [ ] **Step 2: 验证编译**

Run: `cd notification && go build ./repository/cache/...`
Expected: 编译通过

- [ ] **Step 3: 提交**

```bash
git add notification/repository/cache/template.go
git commit -m "feat(notification): 添加 Template Redis 缓存"
```

---

### Task 10: 重写 Notification Repository

**Files:**
- Modify: `notification/repository/notification.go`

- [ ] **Step 1: 重写 Repository 接口和实现**

```go
type NotificationRepository interface {
	Create(ctx context.Context, n domain.Notification) (int64, error)
	BatchCreate(ctx context.Context, ns []domain.Notification) ([]int64, error)
	FindByKeyAndChannel(ctx context.Context, key string, channel domain.Channel) (domain.Notification, error)
	FindByUserId(ctx context.Context, userId int64, offset, limit int) ([]domain.Notification, error)
	FindByUserIdAndGroup(ctx context.Context, userId int64, group domain.NotificationGroup, offset, limit int) ([]domain.Notification, error)
	FindUnreadByUserId(ctx context.Context, userId int64, offset, limit int) ([]domain.Notification, error)
	MarkAsRead(ctx context.Context, userId int64, ids []int64) error
	MarkAllAsRead(ctx context.Context, userId int64) error
	GetUnreadCount(ctx context.Context, userId int64) (domain.UnreadCount, error)
	UpdateStatus(ctx context.Context, id int64, status domain.NotificationStatus) error
	Delete(ctx context.Context, userId int64, id int64) error
	DeleteAll(ctx context.Context, userId int64) error
}
```

CachedNotificationRepository 实现：
- `Create`：DAO Insert → 站内通知时 IncrUnreadCount + PublishSSE
- `BatchCreate`：DAO BatchInsert → 逐条 IncrUnreadCount + PublishSSE
- `GetUnreadCount`：先查缓存 → 未命中查 DAO CountUnreadByGroup → 回填缓存
- `MarkAsRead/MarkAllAsRead`：DAO 更新 → ClearUnreadCount
- toEntity/toDomain 转换函数适配新字段（Key, Channel, TemplateId, TemplateParams, GroupType 等）

- [ ] **Step 2: 验证编译**

Run: `cd notification && go build ./repository/...`
Expected: 编译通过

- [ ] **Step 3: 提交**

```bash
git add notification/repository/notification.go
git commit -m "feat(notification): 重写 Notification Repository 适配新模型"
```

---

### Task 11: 创建 Template Repository

**Files:**
- Create: `notification/repository/template.go`

- [ ] **Step 1: 创建 Template Repository**

```go
type TemplateRepository interface {
	Create(ctx context.Context, t domain.Template) (int64, error)
	Update(ctx context.Context, t domain.Template) error
	FindByTemplateIdAndChannel(ctx context.Context, templateId string, channel domain.Channel) (domain.Template, error)
	FindByChannel(ctx context.Context, channel domain.Channel, offset, limit int) ([]domain.Template, error)
}
```

CachedTemplateRepository 实现：
- `FindByTemplateIdAndChannel`：先查 TemplateCache → 未命中查 DAO → 回填缓存
- `Update`：DAO Update → 删除缓存

- [ ] **Step 2: 验证编译**

Run: `cd notification && go build ./repository/...`
Expected: 编译通过

- [ ] **Step 3: 提交**

```bash
git add notification/repository/template.go
git commit -m "feat(notification): 添加 Template Repository"
```

---

### Task 12: 创建 Transaction Repository

**Files:**
- Create: `notification/repository/transaction.go`

- [ ] **Step 1: 创建 Transaction Repository**

```go
type TransactionRepository interface {
	Create(ctx context.Context, t domain.Transaction) (int64, error)
	FindByKey(ctx context.Context, key string) (domain.Transaction, error)
	UpdateStatus(ctx context.Context, key string, status domain.TransactionStatus) error
	FindPreparedTimeout(ctx context.Context, limit int) ([]domain.Transaction, error)
	IncrRetryCount(ctx context.Context, id int64, nextCheckTime int64) error
}
```

直接封装 TransactionDAO，无缓存层。toDomain/toEntity 转换。

- [ ] **Step 2: 验证编译**

Run: `cd notification && go build ./repository/...`
Expected: 编译通过

- [ ] **Step 3: 提交**

```bash
git add notification/repository/transaction.go
git commit -m "feat(notification): 添加 Transaction Repository"
```

---

## Chunk 3: Channel 层 + Service 层

### Task 13: 创建 ChannelSender 接口和 InAppSender

**Files:**
- Create: `notification/channel/types.go`
- Create: `notification/channel/inapp.go`

- [ ] **Step 1: 创建 ChannelSender 接口**

```go
// notification/channel/types.go
package channel

import (
	"context"
	"gitee.com/geekbang/webook/notification/domain"
)

type Sender interface {
	Send(ctx context.Context, notification domain.Notification) error
	BatchSend(ctx context.Context, notifications []domain.Notification) error
}
```

- [ ] **Step 2: 创建 InAppSender**

```go
// notification/channel/inapp.go
package channel

import (
	"context"
	"gitee.com/geekbang/webook/notification/domain"
	"gitee.com/geekbang/webook/notification/repository"
)

type InAppSender struct {
	repo repository.NotificationRepository
}

func NewInAppSender(repo repository.NotificationRepository) *InAppSender {
	return &InAppSender{repo: repo}
}

func (s *InAppSender) Send(ctx context.Context, n domain.Notification) error {
	_, err := s.repo.Create(ctx, n)
	return err
}

func (s *InAppSender) BatchSend(ctx context.Context, ns []domain.Notification) error {
	_, err := s.repo.BatchCreate(ctx, ns)
	return err
}
```

- [ ] **Step 3: 验证编译**

Run: `cd notification && go build ./channel/...`
Expected: 编译通过

- [ ] **Step 4: 提交**

```bash
git add notification/channel/types.go notification/channel/inapp.go
git commit -m "feat(notification): 添加 ChannelSender 接口和 InAppSender"
```

---

### Task 14: 创建 SMSSender 和 SMSProvider

**Files:**
- Create: `notification/channel/sms_provider.go`
- Create: `notification/channel/sms.go`

- [ ] **Step 1: 创建 SMSProvider 接口和阿里云实现**

```go
// notification/channel/sms_provider.go
package channel

import "context"

type SMSProvider interface {
	Send(ctx context.Context, tplId string, params map[string]string, numbers ...string) error
}
```

阿里云实现：`AliyunSMSProvider`，将 `map[string]string` 序列化为 JSON 调用阿里云 SDK。
腾讯云实现：`TencentSMSProvider`，将 map 按 key 排序取 value 列表调用腾讯云 SDK。

参考现有 `sms/service/aliyun/service.go` 的实现模式。

- [ ] **Step 2: 创建 SMSSender**

```go
// notification/channel/sms.go
package channel

import (
	"context"
	"fmt"

	"gitee.com/geekbang/webook/notification/domain"
	"gitee.com/geekbang/webook/notification/repository"
)

type SMSSender struct {
	provider SMSProvider
	tplRepo  repository.TemplateRepository
}

func NewSMSSender(provider SMSProvider, tplRepo repository.TemplateRepository) *SMSSender {
	return &SMSSender{provider: provider, tplRepo: tplRepo}
}

func (s *SMSSender) Send(ctx context.Context, n domain.Notification) error {
	tpl, err := s.tplRepo.FindByTemplateIdAndChannel(ctx, n.TemplateId, domain.ChannelSMS)
	if err != nil {
		return fmt.Errorf("查询SMS模板失败: %w", err)
	}
	return s.provider.Send(ctx, tpl.SMSProviderTemplateId, n.TemplateParams, n.Receiver)
}

func (s *SMSSender) BatchSend(ctx context.Context, ns []domain.Notification) error {
	for _, n := range ns {
		if err := s.Send(ctx, n); err != nil {
			return err
		}
	}
	return nil
}
```

- [ ] **Step 3: 验证编译**

Run: `cd notification && go build ./channel/...`
Expected: 编译通过

- [ ] **Step 4: 提交**

```bash
git add notification/channel/sms_provider.go notification/channel/sms.go
git commit -m "feat(notification): 添加 SMSSender 和 SMSProvider"
```

---

### Task 15: 创建 EmailSender（预留）

**Files:**
- Create: `notification/channel/email.go`

- [ ] **Step 1: 创建 EmailSender 预留实现**

```go
// notification/channel/email.go
package channel

import (
	"context"
	"fmt"

	"gitee.com/geekbang/webook/notification/domain"
)

type EmailSender struct{}

func NewEmailSender() *EmailSender {
	return &EmailSender{}
}

func (s *EmailSender) Send(ctx context.Context, n domain.Notification) error {
	return fmt.Errorf("email 渠道暂未实现")
}

func (s *EmailSender) BatchSend(ctx context.Context, ns []domain.Notification) error {
	return fmt.Errorf("email 渠道暂未实现")
}
```

- [ ] **Step 2: 提交**

```bash
git add notification/channel/email.go
git commit -m "feat(notification): 添加 EmailSender 预留实现"
```

---

### Task 16: 创建 TemplateService

**Files:**
- Create: `notification/service/template.go`

- [ ] **Step 1: 创建 TemplateService 接口和实现**

```go
// notification/service/template.go
package service

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"gitee.com/geekbang/webook/notification/domain"
	"gitee.com/geekbang/webook/notification/repository"
)

type TemplateService interface {
	Create(ctx context.Context, tpl domain.Template) (int64, error)
	Update(ctx context.Context, tpl domain.Template) error
	GetByTemplateId(ctx context.Context, templateId string, channel domain.Channel) (domain.Template, error)
	List(ctx context.Context, channel domain.Channel, offset, limit int) ([]domain.Template, error)
	Render(ctx context.Context, templateId string, channel domain.Channel, params map[string]string) (string, error)
}

type templateService struct {
	repo repository.TemplateRepository
}

func NewTemplateService(repo repository.TemplateRepository) TemplateService {
	return &templateService{repo: repo}
}

func (s *templateService) Create(ctx context.Context, tpl domain.Template) (int64, error) {
	return s.repo.Create(ctx, tpl)
}

func (s *templateService) Update(ctx context.Context, tpl domain.Template) error {
	return s.repo.Update(ctx, tpl)
}

func (s *templateService) GetByTemplateId(ctx context.Context, templateId string, channel domain.Channel) (domain.Template, error) {
	return s.repo.FindByTemplateIdAndChannel(ctx, templateId, channel)
}

func (s *templateService) List(ctx context.Context, channel domain.Channel, offset, limit int) ([]domain.Template, error) {
	return s.repo.FindByChannel(ctx, channel, offset, limit)
}

func (s *templateService) Render(ctx context.Context, templateId string, channel domain.Channel, params map[string]string) (string, error) {
	tpl, err := s.repo.FindByTemplateIdAndChannel(ctx, templateId, channel)
	if err != nil {
		return "", fmt.Errorf("查询模板失败: %w", err)
	}
	if tpl.Status == domain.TemplateStatusDisabled {
		return "", fmt.Errorf("模板已禁用: %s", templateId)
	}
	t, err := template.New("notification").Parse(tpl.Content)
	if err != nil {
		return "", fmt.Errorf("解析模板失败: %w", err)
	}
	var buf bytes.Buffer
	if err = t.Execute(&buf, params); err != nil {
		return "", fmt.Errorf("渲染模板失败: %w", err)
	}
	return buf.String(), nil
}
```

- [ ] **Step 2: 验证编译**

Run: `cd notification && go build ./service/...`
Expected: 编译通过

- [ ] **Step 3: 提交**

```bash
git add notification/service/template.go
git commit -m "feat(notification): 添加 TemplateService 模板管理与渲染"
```

---

### Task 17: 重写 NotificationService

**Files:**
- Modify: `notification/service/notification.go`

- [ ] **Step 1: 重写 NotificationService 接口和实现**

```go
type NotificationService interface {
	// 普通发送
	Send(ctx context.Context, n domain.Notification) (int64, error)
	BatchSend(ctx context.Context, n domain.Notification) ([]int64, error) // receivers 在 n.Receivers 中
	// TCC 事务
	Prepare(ctx context.Context, req domain.PrepareRequest) (int64, int64, error)
	Confirm(ctx context.Context, key string) error
	Cancel(ctx context.Context, key string) error
	// 站内通知查询
	List(ctx context.Context, userId int64, offset, limit int) ([]domain.Notification, error)
	ListByGroup(ctx context.Context, userId int64, group domain.NotificationGroup, offset, limit int) ([]domain.Notification, error)
	ListUnread(ctx context.Context, userId int64, offset, limit int) ([]domain.Notification, error)
	MarkAsRead(ctx context.Context, userId int64, ids []int64) error
	MarkAllAsRead(ctx context.Context, userId int64) error
	GetUnreadCount(ctx context.Context, userId int64) (domain.UnreadCount, error)
	Delete(ctx context.Context, userId int64, id int64) error
	DeleteAll(ctx context.Context, userId int64) error
}
```

实现类 `notificationService` 包含：
- `senders map[domain.Channel]channel.Sender` — 渠道路由表
- `repo NotificationRepository`
- `txRepo TransactionRepository`
- `tplSvc TemplateService`

核心方法：
- `Send`：生成 key → 站内通知转换 UserId → 幂等检查 → 模板渲染 → 写 DB → 路由到 Sender
- `BatchSend`：遍历 receivers，为每个 receiver 创建 Notification 并调用 Send
- `Prepare`：本地事务写 notifications(Init) + transactions(Prepared)
- `Confirm`：查事务状态 → 本地事务更新 Confirmed+Sending → DB提交后执行外部发送 → 更新 Sent/Failed
- `Cancel`：查事务状态 → 本地事务更新 Cancelled+Failed

- [ ] **Step 2: 验证编译**

Run: `cd notification && go build ./service/...`
Expected: 编译通过

- [ ] **Step 3: 提交**

```bash
git add notification/service/notification.go
git commit -m "feat(notification): 重写 NotificationService 支持多渠道和 TCC"
```

---

## Chunk 4: gRPC Server + 事务回查

### Task 18: 重写 gRPC Server

**Files:**
- Modify: `notification/grpc/server.go`

- [ ] **Step 1: 重写 gRPC server 适配 v2 proto**

导入新的 `notificationv2` 生成代码包。实现所有 18 个 RPC 方法：
- `Send`、`BatchSend` — 调用 NotificationService
- `Prepare`、`Confirm`、`Cancel` — TCC 事务
- `CreateTemplate`、`UpdateTemplate`、`GetTemplate`、`ListTemplates` — 模板管理
- `ListNotifications`、`ListByGroup`、`ListUnread` — 查询
- `MarkAsRead`、`MarkAllAsRead` — 已读
- `GetUnreadCount` — 未读统计（返回 by_group）
- `Delete`、`DeleteAll` — 删除

每个方法做 proto message ↔ domain model 的转换。错误处理按 spec 中的 gRPC 错误码映射表返回。

- [ ] **Step 2: 验证编译**

Run: `cd notification && go build ./grpc/...`
Expected: 编译通过

- [ ] **Step 3: 提交**

```bash
git add notification/grpc/server.go
git commit -m "feat(notification): 重写 gRPC server 适配 v2 proto"
```

---

### Task 19: 创建 CheckBackScheduler

**Files:**
- Create: `notification/scheduler/checkback.go`

- [ ] **Step 1: 创建回查定时任务**

```go
// notification/scheduler/checkback.go
package scheduler

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"

	"gitee.com/geekbang/webook/notification/domain"
	"gitee.com/geekbang/webook/notification/repository"
	"gitee.com/geekbang/webook/notification/service"
	"gitee.com/geekbang/webook/pkg/logger"

	notificationv2 "gitee.com/geekbang/webook/api/proto/gen/notification/v2"
)

type CheckBackScheduler struct {
	txRepo        repository.TransactionRepository
	svc           service.NotificationService
	etcdClient    *clientv3.Client
	l             logger.LoggerV1
	maxRetry      int
	scanInterval  time.Duration
	retryInterval time.Duration
}

func NewCheckBackScheduler(
	txRepo repository.TransactionRepository,
	svc service.NotificationService,
	etcdClient *clientv3.Client,
	l logger.LoggerV1,
) *CheckBackScheduler {
	return &CheckBackScheduler{
		txRepo:        txRepo,
		svc:           svc,
		etcdClient:    etcdClient,
		l:             l,
		maxRetry:      5,
		scanInterval:  10 * time.Second,
		retryInterval: 10 * time.Second,
	}
}

func (s *CheckBackScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.scanInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.scan(ctx)
		}
	}
}

func (s *CheckBackScheduler) scan(ctx context.Context) {
	txs, err := s.txRepo.FindPreparedTimeout(ctx, 100)
	if err != nil {
		s.l.Error("回查扫描失败", logger.Error(err))
		return
	}
	for _, tx := range txs {
		s.checkOne(ctx, tx)
	}
}

func (s *CheckBackScheduler) checkOne(ctx context.Context, tx domain.Transaction) {
	if tx.RetryCount >= tx.MaxRetry {
		s.l.Warn("回查超过最大次数，强制取消",
			logger.String("key", tx.Key),
			logger.String("biz_id", tx.BizId))
		_ = s.svc.Cancel(ctx, tx.Key)
		return
	}

	// 从 ETCD 发现 TransactionChecker 服务
	prefix := fmt.Sprintf("/services/transaction-checker/%s/", tx.BizId)
	resp, err := s.etcdClient.Get(ctx, prefix, clientv3.WithPrefix(), clientv3.WithLimit(1))
	if err != nil || len(resp.Kvs) == 0 {
		s.l.Warn("找不到回查服务",
			logger.String("biz_id", tx.BizId),
			logger.Error(err))
		nextCheck := time.Now().UnixMilli() + s.retryInterval.Milliseconds()*int64(tx.RetryCount+1)
		_ = s.txRepo.IncrRetryCount(ctx, tx.Id, nextCheck)
		return
	}

	addr := string(resp.Kvs[0].Value)
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if err != nil {
		s.l.Error("连接回查服务失败", logger.String("addr", addr), logger.Error(err))
		nextCheck := time.Now().UnixMilli() + s.retryInterval.Milliseconds()*int64(tx.RetryCount+1)
		_ = s.txRepo.IncrRetryCount(ctx, tx.Id, nextCheck)
		return
	}
	defer conn.Close()

	client := notificationv2.NewTransactionCheckerClient(conn)
	result, err := client.CheckTransaction(ctx, &notificationv2.CheckTransactionRequest{Key: tx.Key})
	if err != nil {
		s.l.Error("回查调用失败", logger.String("key", tx.Key), logger.Error(err))
		nextCheck := time.Now().UnixMilli() + s.retryInterval.Milliseconds()*int64(tx.RetryCount+1)
		_ = s.txRepo.IncrRetryCount(ctx, tx.Id, nextCheck)
		return
	}

	switch result.Action {
	case notificationv2.TransactionAction_TRANSACTION_ACTION_COMMIT:
		_ = s.svc.Confirm(ctx, tx.Key)
	case notificationv2.TransactionAction_TRANSACTION_ACTION_ROLLBACK:
		_ = s.svc.Cancel(ctx, tx.Key)
	case notificationv2.TransactionAction_TRANSACTION_ACTION_PENDING:
		nextCheck := time.Now().UnixMilli() + s.retryInterval.Milliseconds()*int64(tx.RetryCount+1)
		_ = s.txRepo.IncrRetryCount(ctx, tx.Id, nextCheck)
	}
}
```

- [ ] **Step 2: 验证编译**

Run: `cd notification && go build ./scheduler/...`
Expected: 编译通过

- [ ] **Step 3: 提交**

```bash
git add notification/scheduler/checkback.go
git commit -m "feat(notification): 添加 CheckBackScheduler 事务回查"
```

---

## Chunk 5: Kafka 消费者适配 + IoC + Wire + 启动

### Task 20: 适配 Kafka 消费者

**Files:**
- Modify: `notification/events/consumer.go`

- [ ] **Step 1: 重写所有 5 个消费者**

所有消费者改为注入 `service.NotificationService` 替代 `repository.NotificationRepository`。

每个消费者的 Consume 方法改为构造 `domain.Notification` 并调用 `svc.Send`：
- 设置 `Key`（幂等键，如 `"like:{liker}:{targetId}"`）
- 设置 `Channel = domain.ChannelInApp`
- 设置 `Receiver = strconv.FormatInt(userId, 10)`
- 设置 `TemplateId`（需要预置对应模板，如 `"like_notification"`）
- 设置 `TemplateParams`
- 设置 `GroupType`（按映射表：Like/Collect→Interaction, Comment/Reply→Reply, Follow→Follow）
- 设置 `SourceId`, `SourceName`, `TargetId`, `TargetType`, `TargetTitle`

注意：events/types.go 中的事件结构体保持不变，不需要改动。

- [ ] **Step 2: 验证编译**

Run: `cd notification && go build ./events/...`
Expected: 编译通过

- [ ] **Step 3: 提交**

```bash
git add notification/events/consumer.go
git commit -m "feat(notification): 适配 Kafka 消费者调用 NotificationService.Send"
```

---

### Task 21: 创建 IoC 初始化文件

**Files:**
- Create: `notification/ioc/channel.go`
- Create: `notification/ioc/scheduler.go`
- Create: `notification/ioc/template.go`

- [ ] **Step 1: 创建 channel.go**

```go
// notification/ioc/channel.go
package ioc

import (
	"gitee.com/geekbang/webook/notification/channel"
	"gitee.com/geekbang/webook/notification/domain"
	"gitee.com/geekbang/webook/notification/repository"
)

func InitChannelSenders(
	inApp *channel.InAppSender,
	sms *channel.SMSSender,
	email *channel.EmailSender,
) map[domain.Channel]channel.Sender {
	return map[domain.Channel]channel.Sender{
		domain.ChannelInApp: inApp,
		domain.ChannelSMS:   sms,
		domain.ChannelEmail: email,
	}
}

func InitSMSProvider() channel.SMSProvider {
	// 根据配置选择阿里云或腾讯云，这里先返回阿里云实现
	// 实际从 Viper 读取配置决定
	return channel.NewAliyunSMSProvider(/* config params */)
}
```

- [ ] **Step 2: 创建 scheduler.go**

```go
// notification/ioc/scheduler.go
package ioc

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"github.com/spf13/viper"

	"gitee.com/geekbang/webook/notification/repository"
	"gitee.com/geekbang/webook/notification/scheduler"
	"gitee.com/geekbang/webook/notification/service"
	"gitee.com/geekbang/webook/pkg/logger"
)

func InitETCDClient() *clientv3.Client {
	type Config struct {
		Addrs []string `yaml:"addrs"`
	}
	var cfg Config
	err := viper.UnmarshalKey("etcd", &cfg)
	if err != nil {
		panic(err)
	}
	client, err := clientv3.New(clientv3.Config{Endpoints: cfg.Addrs})
	if err != nil {
		panic(err)
	}
	return client
}

func InitCheckBackScheduler(
	txRepo repository.TransactionRepository,
	svc service.NotificationService,
	etcdClient *clientv3.Client,
	l logger.LoggerV1,
) *scheduler.CheckBackScheduler {
	return scheduler.NewCheckBackScheduler(txRepo, svc, etcdClient, l)
}
```

- [ ] **Step 3: 创建 template.go**

```go
// notification/ioc/template.go
package ioc

// 如果模板服务需要特殊初始化逻辑放在这里
// 目前 TemplateService 通过 Wire 直接注入，无需额外 IoC
```

- [ ] **Step 4: 验证编译**

Run: `cd notification && go build ./ioc/...`
Expected: 编译通过

- [ ] **Step 5: 提交**

```bash
git add notification/ioc/channel.go notification/ioc/scheduler.go notification/ioc/template.go
git commit -m "feat(notification): 添加 IoC 初始化（channel, scheduler, template）"
```

---

### Task 22: 更新 Wire 依赖注入

**Files:**
- Modify: `notification/wire.go`
- Modify: `notification/app.go`

- [ ] **Step 1: 更新 wire.go**

```go
var thirdPartySet = wire.NewSet(
	ioc.InitDB,
	ioc.InitLogger,
	ioc.InitRedis,
	ioc.InitKafka,
	ioc.InitSyncProducer,
	ioc.InitETCDClient,
)

var templateSet = wire.NewSet(
	service.NewTemplateService,
	repository.NewCachedTemplateRepository,
	dao.NewGORMTemplateDAO,
	cache.NewRedisTemplateCache,
)

var channelSet = wire.NewSet(
	channel.NewInAppSender,
	channel.NewSMSSender,
	channel.NewEmailSender,
	ioc.InitSMSProvider,
	ioc.InitChannelSenders,
)

var notificationSet = wire.NewSet(
	grpc.NewNotificationServiceServer,
	service.NewNotificationService,
	repository.NewCachedNotificationRepository,
	repository.NewTransactionRepository,
	dao.NewGORMNotificationDAO,
	dao.NewGORMTransactionDAO,
	cache.NewRedisNotificationCache,
)

var schedulerSet = wire.NewSet(
	ioc.InitCheckBackScheduler,
)

var consumerSet = wire.NewSet(
	events.NewNotificationEventConsumer,
	events.NewLikeEventConsumer,
	events.NewCollectEventConsumer,
	events.NewCommentEventConsumer,
	events.NewFollowEventConsumer,
	ioc.NewConsumers,
)

func InitApp() *App {
	wire.Build(
		thirdPartySet,
		templateSet,
		channelSet,
		notificationSet,
		schedulerSet,
		consumerSet,
		ioc.InitGRPCServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
```

- [ ] **Step 2: 更新 app.go**

```go
// notification/app.go
package main

import (
	"gitee.com/geekbang/webook/notification/scheduler"
	"gitee.com/geekbang/webook/pkg/grpcx"
	"gitee.com/geekbang/webook/pkg/saramax"
)

type App struct {
	server    *grpcx.Server
	consumers []saramax.Consumer
	scheduler *scheduler.CheckBackScheduler
}
```

- [ ] **Step 3: 更新 main.go 启动 scheduler**

```go
// 在 main() 中，consumers 启动后加入：
go app.scheduler.Start(context.Background())
```

- [ ] **Step 4: 运行 Wire 生成**

Run: `cd notification && wire`
Expected: `wire_gen.go` 重新生成，无报错

- [ ] **Step 5: 验证整体编译**

Run: `cd notification && go build .`
Expected: 编译通过

- [ ] **Step 6: 提交**

```bash
git add notification/wire.go notification/wire_gen.go notification/app.go notification/main.go
git commit -m "feat(notification): 更新 Wire DI 和启动流程"
```

---

## Chunk 6: SQL 迁移 + 配置 + 最终验证

### Task 23: 更新 SQL 初始化脚本

**Files:**
- Modify: `script/mysql/init.sql`

- [ ] **Step 1: 在 init.sql 中添加三张新表的建表语句**

按 spec 中的 SQL 定义，添加 `notifications`（新版）、`notification_transactions`、`notification_templates` 三张表。

- [ ] **Step 2: 提交**

```bash
git add script/mysql/init.sql
git commit -m "feat(notification): 添加新版通知表、事务表、模板表 SQL"
```

---

### Task 24: 更新配置文件

**Files:**
- Modify: `notification/config/docker.yaml`（或 `dev.yaml`）

- [ ] **Step 1: 添加 ETCD 和 SMS 配置项**

```yaml
etcd:
  addrs:
    - "localhost:2379"
sms:
  provider: "aliyun"  # aliyun | tencent
  aliyun:
    accessKeyId: ""
    accessKeySecret: ""
    signName: ""
    endpoint: ""
  tencent:
    secretId: ""
    secretKey: ""
    appId: ""
    signName: ""
```

- [ ] **Step 2: 提交**

```bash
git add notification/config/
git commit -m "feat(notification): 添加 ETCD 和 SMS 配置项"
```

---

### Task 25: 最终编译和验证

- [ ] **Step 1: 完整编译 notification 模块**

Run: `cd notification && go build .`
Expected: 编译通过，无错误

- [ ] **Step 2: 检查所有包编译**

Run: `go build ./...` （项目根目录）
Expected: 编译通过（可能需要调整 BFF 层引用）

- [ ] **Step 3: 最终提交**

```bash
git add .
git commit -m "feat(notification): 通知服务重构完成 - 统一网关/模板管理/TCC事务/回查机制"
```

---

## 依赖关系

```
Task 1 (types.go)
  → Task 2 (notification domain) → Task 3 (template/transaction domain)
    → Task 4 (proto v2)
      → Task 5 (notification DAO) → Task 6 (template DAO) → Task 7 (transaction DAO)
        → Task 8 (notification cache) → Task 9 (template cache)
          → Task 10 (notification repo) → Task 11 (template repo) → Task 12 (transaction repo)
            → Task 13 (channel types + inapp) → Task 14 (sms) → Task 15 (email)
              → Task 16 (template service)
                → Task 17 (notification service)
                  → Task 18 (gRPC server)
                  → Task 19 (checkback scheduler)
                  → Task 20 (kafka consumers)
                    → Task 21 (IoC) → Task 22 (Wire + main)
                      → Task 23 (SQL) → Task 24 (config) → Task 25 (final verify)
```
