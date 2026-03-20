# 通知服务重构设计文档

## 概述

重构现有通知微服务，从仅支持站内通知扩展为统一通知网关，支持站内通知、SMS、Email（预留）三个渠道，提供模板管理、TCC 分布式事务和事务回查机制。

## 背景

### 现有架构

- 分层架构：Domain → Service → Repository (Cache + DAO)
- 5 个 Kafka 消费者（like, collect, comment, follow, notification 事件）
- MySQL 持久化 + Redis 缓存（未读计数）
- SSE 实时推送（通过 Redis Pub/Sub）
- gRPC API（10 个方法）+ BFF REST 层（7 个端点）

### 当前局限

- 只支持站内通知，无 SMS/Email 渠道
- 无通知模板机制
- 无分布式事务保障
- 站内通知类型硬编码（Like/Collect/Comment 等），不够灵活
- 独立 SMS 微服务需要后续废弃

### 设计目标

1. 统一通知网关：站内、SMS、Email 全部内置在通知服务中
2. SMS 逻辑在通知服务内重写，后续废弃独立 sms 微服务
3. Email 预留接口，暂不实现
4. 模板管理：通知服务管理模板 CRUD，调用方传 template_id + params
5. TCC 事务：提供 Prepare/Confirm/Cancel 接口，调用方自选是否使用
6. 事务回查：定时扫描超时 prepared 记录，通过 ETCD 服务发现回查业务方
7. 新增触发场景：交易支付、系统运维相关通知

## 架构设计

### 整体架构：渠道策略模式

通知服务内部用 Strategy 模式抽象渠道层，每个渠道实现统一的 `ChannelSender` 接口。

```
调用方 → gRPC API → NotificationService
                         │
                    ┌─────┼─────┐
                    ▼     ▼     ▼
                 InApp   SMS   Email(预留)
                  │       │
                 DAO   阿里云/腾讯云
                  │
              Redis SSE
```

### 内部分层

```
gRPC Server
    │
NotificationService (业务编排层)
    │
    ├── TemplateService (模板渲染)
    │       └── TemplateRepository → DAO + Cache
    │
    ├── ChannelSender (渠道策略接口)
    │       ├── InAppSender   → NotificationRepository → DAO + Cache + SSE Pub
    │       ├── SMSSender     → 阿里云/腾讯云 SDK
    │       └── EmailSender   → 预留，返回未实现错误
    │
    ├── TransactionManager (TCC 状态管理)
    │       └── TransactionRepository → DAO
    │
    └── CheckBackScheduler (事务回查定时任务)
            └── ETCD → TransactionChecker (业务方)
```

各层职责：

- **NotificationService**：核心编排层。Send 根据 channel 选择 ChannelSender，调用 TemplateService 渲染内容后发送。Prepare/Confirm/Cancel 管理 TCC 流程。
- **ChannelSender 接口**：每个渠道实现 Send/BatchSend，NotificationService 通过 channel 字段路由。
- **TemplateService**：模板 CRUD + 渲染（查缓存 → DB → 回填缓存 → text/template 变量替换）。
- **TransactionManager**：TCC 状态流转管理。
- **CheckBackScheduler**：定时扫描超时 prepared 记录，通过 ETCD 服务发现回查业务方。

## gRPC API 设计

### 枚举定义

```protobuf
enum Channel {
  CHANNEL_UNSPECIFIED = 0;
  CHANNEL_IN_APP = 1;
  CHANNEL_SMS = 2;
  CHANNEL_EMAIL = 3;
}

enum SendStrategy {
  SEND_STRATEGY_UNSPECIFIED = 0;
  SEND_STRATEGY_IMMEDIATE = 1;
  SEND_STRATEGY_SCHEDULED = 2;
}

enum NotificationStatus {
  STATUS_UNSPECIFIED = 0;
  STATUS_INIT = 1;
  STATUS_SENDING = 2;
  STATUS_SENT = 3;
  STATUS_FAILED = 4;
}

enum TransactionStatus {
  TRANSACTION_STATUS_UNSPECIFIED = 0;
  TRANSACTION_STATUS_PREPARED = 1;
  TRANSACTION_STATUS_CONFIRMED = 2;
  TRANSACTION_STATUS_CANCELLED = 3;
}

enum TransactionAction {
  TRANSACTION_ACTION_UNSPECIFIED = 0;
  TRANSACTION_ACTION_COMMIT = 1;
  TRANSACTION_ACTION_ROLLBACK = 2;
  TRANSACTION_ACTION_PENDING = 3;
}

enum NotificationGroup {
  NOTIFICATION_GROUP_UNSPECIFIED = 0;
  NOTIFICATION_GROUP_INTERACTION = 1;
  NOTIFICATION_GROUP_REPLY = 2;
  NOTIFICATION_GROUP_MENTION = 3;
  NOTIFICATION_GROUP_FOLLOW = 4;
  NOTIFICATION_GROUP_SYSTEM = 5;
}
```

### 核心消息

```protobuf
message Notification {
  string key = 1;
  repeated string receivers = 2;
  Channel channel = 3;
  string template_id = 4;
  map<string, string> template_params = 5;
  SendStrategy strategy = 6;
  string receiver = 7;
}
```

### 通知服务接口

```protobuf
service NotificationService {
  // 普通发送
  rpc Send(SendRequest) returns (SendResponse);
  rpc BatchSend(BatchSendRequest) returns (BatchSendResponse);

  // TCC 事务
  rpc Prepare(PrepareRequest) returns (PrepareResponse);
  rpc Confirm(ConfirmRequest) returns (ConfirmResponse);
  rpc Cancel(CancelRequest) returns (CancelResponse);

  // 模板管理
  rpc CreateTemplate(CreateTemplateRequest) returns (CreateTemplateResponse);
  rpc UpdateTemplate(UpdateTemplateRequest) returns (UpdateTemplateResponse);
  rpc GetTemplate(GetTemplateRequest) returns (GetTemplateResponse);
  rpc ListTemplates(ListTemplatesRequest) returns (ListTemplatesResponse);

  // 站内通知查询（保留现有）
  rpc ListNotifications(ListNotificationsRequest) returns (ListNotificationsResponse);
  rpc MarkAsRead(MarkAsReadRequest) returns (MarkAsReadResponse);
  rpc MarkAllAsRead(MarkAllAsReadRequest) returns (MarkAllAsReadResponse);
  rpc GetUnreadCount(GetUnreadCountRequest) returns (GetUnreadCountResponse);
}
```

### TCC 事务相关消息

```protobuf
message PrepareRequest {
  Notification notification = 1;
  string biz_id = 2;
  int64 check_back_timeout_ms = 3;
}

message ConfirmRequest {
  string key = 1;
}

message CancelRequest {
  string key = 1;
}
```

### 事务回查接口（业务方实现，注册到 ETCD）

```protobuf
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

## 数据模型

### 通知记录表 notifications

```sql
CREATE TABLE notifications (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    key_field VARCHAR(256) NOT NULL DEFAULT '',
    biz_id VARCHAR(64) NOT NULL DEFAULT '',
    channel TINYINT NOT NULL DEFAULT 0,
    receiver VARCHAR(256) NOT NULL DEFAULT '',
    template_id VARCHAR(128) NOT NULL DEFAULT '',
    template_params JSON,
    content TEXT,
    status TINYINT NOT NULL DEFAULT 0,
    strategy TINYINT NOT NULL DEFAULT 1,
    -- 站内通知专用
    group_type TINYINT NOT NULL DEFAULT 0,
    source_id BIGINT NOT NULL DEFAULT 0,
    source_name VARCHAR(128) NOT NULL DEFAULT '',
    target_id BIGINT NOT NULL DEFAULT 0,
    target_type VARCHAR(64) NOT NULL DEFAULT '',
    target_title VARCHAR(256) NOT NULL DEFAULT '',
    is_read TINYINT NOT NULL DEFAULT 0,
    ctime BIGINT NOT NULL DEFAULT 0,
    utime BIGINT NOT NULL DEFAULT 0,
    UNIQUE KEY uk_key_channel (key_field, channel),
    INDEX idx_receiver_channel (receiver, channel, ctime DESC),
    INDEX idx_receiver_unread (receiver, is_read, ctime DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 事务通知表 notification_transactions

```sql
CREATE TABLE notification_transactions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    notification_id BIGINT NOT NULL DEFAULT 0,
    key_field VARCHAR(256) NOT NULL DEFAULT '',
    biz_id VARCHAR(64) NOT NULL DEFAULT '',
    status TINYINT NOT NULL DEFAULT 0,
    check_back_timeout_ms BIGINT NOT NULL DEFAULT 30000,
    next_check_time BIGINT NOT NULL DEFAULT 0,
    retry_count INT NOT NULL DEFAULT 0,
    max_retry INT NOT NULL DEFAULT 5,
    ctime BIGINT NOT NULL DEFAULT 0,
    utime BIGINT NOT NULL DEFAULT 0,
    UNIQUE KEY uk_notification_id (notification_id),
    UNIQUE KEY uk_key (key_field),
    INDEX idx_status_check (status, next_check_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 模板表 notification_templates

```sql
CREATE TABLE notification_templates (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    template_id VARCHAR(128) NOT NULL DEFAULT '',
    channel TINYINT NOT NULL DEFAULT 0,
    name VARCHAR(256) NOT NULL DEFAULT '',
    content TEXT NOT NULL,
    description VARCHAR(512) NOT NULL DEFAULT '',
    status TINYINT NOT NULL DEFAULT 1,
    sms_sign VARCHAR(64) NOT NULL DEFAULT '',
    sms_provider_template_id VARCHAR(128) NOT NULL DEFAULT '',
    ctime BIGINT NOT NULL DEFAULT 0,
    utime BIGINT NOT NULL DEFAULT 0,
    UNIQUE KEY uk_template_channel (template_id, channel),
    INDEX idx_channel_status (channel, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## 枚举定义（Go Domain 层）

```go
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
    SendStrategyScheduled SendStrategy = 2
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

## 模板渲染

模板内容使用 `{{.变量名}}` 占位符（Go 标准 `text/template` 语法）。

```
模板内容: "您的订单{{.orderNo}}已支付成功，金额{{.amount}}元"
参数: {"orderNo": "2024001", "amount": "99.00"}
渲染结果: "您的订单2024001已支付成功，金额99.00元"
```

TemplateService 接口：

```go
type TemplateService interface {
    Create(ctx context.Context, tpl domain.Template) (int64, error)
    Update(ctx context.Context, tpl domain.Template) error
    GetByTemplateId(ctx context.Context, templateId string, channel domain.Channel) (domain.Template, error)
    List(ctx context.Context, channel domain.Channel, offset, limit int) ([]domain.Template, error)
    Render(ctx context.Context, templateId string, channel domain.Channel, params map[string]string) (string, error)
}
```

渲染流程：查缓存 → 未命中查 DB → 回填缓存 → text/template 变量替换 → 返回渲染后内容。

## SMS 发送

在通知服务内部重写 SMS 发送逻辑，复用现有 sms 微服务中的阿里云/腾讯云 SDK 封装思路：

```go
type SMSSender struct {
    provider SMSProvider
}

type SMSProvider interface {
    Send(ctx context.Context, tplId string, params []string, numbers ...string) error
}
```

实现类：AliyunSMSProvider、TencentSMSProvider。

SMS 渠道特殊点：
- `template_id` 对应模板表中的 `sms_provider_template_id`（三方平台审核过的模板 ID）
- `receiver` 是手机号
- 通知服务侧的模板内容仅用于记录/展示，实际发送使用三方平台的模板 ID + 参数

## ChannelSender 接口

```go
type ChannelSender interface {
    Send(ctx context.Context, notification domain.Notification) error
    BatchSend(ctx context.Context, notifications []domain.Notification) error
}
```

三个实现：
- **InAppSender**：写 DB + 缓存 + SSE 推送
- **SMSSender**：调用阿里云/腾讯云 SMS SDK
- **EmailSender**：预留，返回 ErrNotImplemented

NotificationService 路由逻辑：

```go
func (s *notificationService) Send(ctx context.Context, n domain.Notification) error {
    // 1. 幂等检查（key + channel）
    // 2. 模板渲染
    content, err := s.templateSvc.Render(ctx, n.TemplateId, n.Channel, n.TemplateParams)
    // 3. 填充渲染后内容
    n.Content = content
    // 4. 路由到对应渠道发送
    sender := s.senders[n.Channel]
    return sender.Send(ctx, n)
}
```

## TCC 事务流程

### 状态流转

```
普通发送:  Init → Sending → Sent/Failed

TCC 发送:  Init ──→ (transactions: Prepared)
               ├── Confirm ──→ Sending → Sent/Failed (transactions: Confirmed)
               └── Cancel  ──→ Failed (transactions: Cancelled)

回查补偿:  Prepared 超时 ──→ CheckBack
               ├── Commit   → Confirm 流程
               ├── Rollback → Cancel 流程
               └── Pending  → 延长超时，下次再查
```

### Prepare

```go
func (s *notificationService) Prepare(ctx context.Context, req domain.PrepareRequest) error {
    // 1. 幂等检查（key + channel）
    // 2. 写 notifications 表，status = Init
    // 3. 写 notification_transactions 表，status = Prepared
    //    设置 next_check_time = now + check_back_timeout_ms
    // 4. 两张表在同一个本地事务中写入
    return nil
}
```

### Confirm

```go
func (s *notificationService) Confirm(ctx context.Context, key string) error {
    // 1. 查 notification_transactions，校验 status == Prepared
    // 2. 本地事务：
    //    - 更新 transactions status = Confirmed
    //    - 模板渲染 + 填充 notifications.content
    //    - 路由到 ChannelSender 发送
    //    - 更新 notifications status = Sent / Failed
    return nil
}
```

### Cancel

```go
func (s *notificationService) Cancel(ctx context.Context, key string) error {
    // 1. 查 notification_transactions，校验 status == Prepared
    // 2. 本地事务：
    //    - 更新 transactions status = Cancelled
    //    - 更新 notifications status = Failed
    return nil
}
```

## 事务回查机制

```go
type CheckBackScheduler struct {
    txRepo        TransactionRepository
    svc           NotificationService
    etcdClient    clientv3.Client
    maxRetry      int           // 最大回查次数，默认 5
    scanInterval  time.Duration // 扫描间隔，默认 10s
    retryInterval time.Duration // 重试间隔递增基数，默认 10s
}
```

回查流程：
1. 定时扫描 `status=Prepared AND next_check_time < now` 的记录
2. 遍历每条记录：
   - `retry_count >= maxRetry` → 强制 Cancel，记录告警日志
   - 通过 `biz_id` 从 ETCD 发现 TransactionChecker 服务
   - 调用 `CheckTransaction(key)`
   - 根据返回结果：
     - Commit → 调用 Confirm(key)
     - Rollback → 调用 Cancel(key)
     - Pending → retry_count++，next_check_time = now + retryInterval * retry_count（递增退避）

退避策略：基数 10s，第 1 次 10s 后重试，第 2 次 20s，第 3 次 30s...最多 5 次后强制 Cancel。

## 站内通知分组

参考 B站、小红书、知乎、微博等平台的消息中心设计，站内通知按消息分组（Tab）维度分类，而非硬编码具体业务动作：

| 分组 | 说明 | 示例 |
|------|------|------|
| 互动消息 | 赞、收藏、投币等互动行为 | "xxx 赞了你的文章" |
| 回复我的 | 评论、回复 | "xxx 回复了你的评论" |
| @我的 | 被提及 | "xxx 在评论中提到了你" |
| 关注 | 关注行为 | "xxx 关注了你" |
| 系统通知 | 系统/运维/交易通知 | "您的订单已支付成功" |

新增业务动作不需要修改通知服务枚举，业务方在发送时指定 group_type 即可。

## 目录结构

```
notification/
├── main.go
├── app.go
├── wire.go
├── wire_gen.go
├── config/
├── domain/
│   ├── notification.go
│   ├── template.go
│   ├── transaction.go
│   └── types.go
├── grpc/
│   └── server.go
├── service/
│   ├── notification.go
│   └── template.go
├── repository/
│   ├── notification.go
│   ├── template.go
│   ├── transaction.go
│   ├── dao/
│   │   ├── init.go
│   │   ├── notification.go
│   │   ├── template.go
│   │   └── transaction.go
│   └── cache/
│       ├── notification.go
│       └── template.go
├── channel/
│   ├── types.go
│   ├── inapp.go
│   ├── sms.go
│   ├── sms_provider.go
│   └── email.go
├── scheduler/
│   └── checkback.go
├── events/
│   ├── types.go
│   ├── consumer.go
│   └── producer.go
└── ioc/
    ├── db.go
    ├── redis.go
    ├── kafka.go
    ├── grpc.go
    ├── logger.go
    ├── channel.go
    ├── scheduler.go
    └── template.go
```

## 保留的现有能力

- Kafka 消费者（like, collect, comment, follow, notification 事件）继续工作，消费者内部适配新的 group_type 字段
- SSE 实时推送（Redis Pub/Sub）保持不变
- BFF 层 handler 和 SSE Hub 保持不变
- Redis 未读计数缓存保持不变
