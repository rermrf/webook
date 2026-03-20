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
  SEND_STRATEGY_SCHEDULED = 2;  // 预留
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
  NOTIFICATION_GROUP_INTERACTION = 1;  // 互动消息
  NOTIFICATION_GROUP_REPLY = 2;        // 回复我的
  NOTIFICATION_GROUP_MENTION = 3;      // @我的
  NOTIFICATION_GROUP_FOLLOW = 4;       // 关注
  NOTIFICATION_GROUP_SYSTEM = 5;       // 系统通知
}
```

### 核心消息

```protobuf
// 通知消息体（调用方构造）
message Notification {
  // 业务方幂等键（必填，用于去重。格式建议：{biz}:{action}:{id}，如 "payment:success:12345"）
  string key = 1;
  // 批量接收者（批量发送时使用，与 receiver 二选一）
  // 站内通知传用户ID字符串，SMS传手机号，Email传邮箱
  repeated string receivers = 2;
  // 发送渠道
  Channel channel = 3;
  // 模板ID
  string template_id = 4;
  // 模板参数
  map<string, string> template_params = 5;
  // 发送策略
  SendStrategy strategy = 6;
  // 单个接收者（单发时使用，与 receivers 二选一）
  string receiver = 7;
  // 站内通知分组（仅 channel=IN_APP 时需要）
  NotificationGroup group_type = 8;
  // 站内通知扩展字段（仅 channel=IN_APP 时需要）
  int64 source_id = 9;
  string source_name = 10;
  int64 target_id = 11;
  string target_type = 12;
  string target_title = 13;
}
```

**receiver 字段说明**：
- `receiver` 用于单发，`receivers` 用于批量，二选一。如果两个都传，以 `receivers` 为准。
- 站内通知：传用户 ID 的字符串形式（如 "12345"），服务内部转换为 int64 使用。
- SMS：传手机号（如 "13800138000"）。
- Email：传邮箱地址（如 "user@example.com"）。

### 通知服务接口

```protobuf
service NotificationService {
  // ===== 普通发送 =====
  rpc Send(SendRequest) returns (SendResponse);
  rpc BatchSend(BatchSendRequest) returns (BatchSendResponse);

  // ===== TCC 事务 =====
  rpc Prepare(PrepareRequest) returns (PrepareResponse);
  rpc Confirm(ConfirmRequest) returns (ConfirmResponse);
  rpc Cancel(CancelRequest) returns (CancelResponse);

  // ===== 模板管理 =====
  rpc CreateTemplate(CreateTemplateRequest) returns (CreateTemplateResponse);
  rpc UpdateTemplate(UpdateTemplateRequest) returns (UpdateTemplateResponse);
  rpc GetTemplate(GetTemplateRequest) returns (GetTemplateResponse);
  rpc ListTemplates(ListTemplatesRequest) returns (ListTemplatesResponse);

  // ===== 站内通知查询 =====
  rpc ListNotifications(ListNotificationsRequest) returns (ListNotificationsResponse);
  rpc ListByGroup(ListByGroupRequest) returns (ListByGroupResponse);
  rpc ListUnread(ListUnreadRequest) returns (ListUnreadResponse);
  rpc MarkAsRead(MarkAsReadRequest) returns (MarkAsReadResponse);
  rpc MarkAllAsRead(MarkAllAsReadRequest) returns (MarkAllAsReadResponse);
  rpc GetUnreadCount(GetUnreadCountRequest) returns (GetUnreadCountResponse);
  rpc Delete(DeleteRequest) returns (DeleteResponse);
  rpc DeleteAll(DeleteAllRequest) returns (DeleteAllResponse);
}
```

### 请求/响应消息定义

```protobuf
// ===== 普通发送 =====
message SendRequest {
  Notification notification = 1;
}
message SendResponse {
  int64 notification_id = 1;
}

message BatchSendRequest {
  Notification notification = 1;  // receivers 字段填充多个接收者
}
message BatchSendResponse {
  repeated int64 notification_ids = 1;
}

// ===== TCC 事务 =====
message PrepareRequest {
  Notification notification = 1;
  string biz_id = 2;               // 业务方标识，用于 ETCD 服务发现回查
  int64 check_back_timeout_ms = 3;  // 超时后触发回查，默认 30000ms
}
message PrepareResponse {
  int64 notification_id = 1;
  int64 transaction_id = 2;
}

message ConfirmRequest {
  string key = 1;  // 通过业务幂等键关联
}
message ConfirmResponse {}

message CancelRequest {
  string key = 1;
}
message CancelResponse {}

// ===== 模板管理 =====
message CreateTemplateRequest {
  string template_id = 1;               // 业务方使用的模板编码
  Channel channel = 2;
  string name = 3;
  string content = 4;                   // 模板内容，含 {{.变量名}} 占位符
  string description = 5;
  string sms_sign = 6;                  // SMS专用：短信签名
  string sms_provider_template_id = 7;  // SMS专用：三方平台模板ID
}
message CreateTemplateResponse {
  int64 id = 1;
}

message UpdateTemplateRequest {
  string template_id = 1;
  Channel channel = 2;
  string name = 3;
  string content = 4;
  string description = 5;
  int32 status = 6;                     // 1启用 2禁用
  string sms_sign = 7;
  string sms_provider_template_id = 8;
}
message UpdateTemplateResponse {}

message GetTemplateRequest {
  string template_id = 1;
  Channel channel = 2;
}
message GetTemplateResponse {
  Template template = 1;
}

message ListTemplatesRequest {
  Channel channel = 1;
  int64 offset = 2;
  int64 limit = 3;
}
message ListTemplatesResponse {
  repeated Template templates = 1;
}

message Template {
  int64 id = 1;
  string template_id = 2;
  Channel channel = 3;
  string name = 4;
  string content = 5;
  string description = 6;
  int32 status = 7;
  string sms_sign = 8;
  string sms_provider_template_id = 9;
  int64 ctime = 10;
  int64 utime = 11;
}

// ===== 站内通知查询 =====
message ListNotificationsRequest {
  int64 user_id = 1;
  int64 offset = 2;
  int64 limit = 3;
}
message ListNotificationsResponse {
  repeated NotificationItem notifications = 1;
}

message ListByGroupRequest {
  int64 user_id = 1;
  NotificationGroup group_type = 2;
  int64 offset = 3;
  int64 limit = 4;
}
message ListByGroupResponse {
  repeated NotificationItem notifications = 1;
}

message ListUnreadRequest {
  int64 user_id = 1;
  int64 offset = 2;
  int64 limit = 3;
}
message ListUnreadResponse {
  repeated NotificationItem notifications = 1;
}

message MarkAsReadRequest {
  int64 user_id = 1;
  repeated int64 notification_ids = 2;  // 支持批量标记已读
}
message MarkAsReadResponse {}

message MarkAllAsReadRequest {
  int64 user_id = 1;
}
message MarkAllAsReadResponse {}

message GetUnreadCountRequest {
  int64 user_id = 1;
}
message GetUnreadCountResponse {
  int64 total = 1;
  map<int32, int64> by_group = 2;  // key 为 NotificationGroup 枚举值
}

message DeleteRequest {
  int64 user_id = 1;
  int64 notification_id = 2;
}
message DeleteResponse {}

message DeleteAllRequest {
  int64 user_id = 1;
}
message DeleteAllResponse {}

// 站内通知查询返回的通知条目
message NotificationItem {
  int64 id = 1;
  NotificationGroup group_type = 2;
  int64 source_id = 3;
  string source_name = 4;
  int64 target_id = 5;
  string target_type = 6;
  string target_title = 7;
  string content = 8;
  bool is_read = 9;
  int64 ctime = 10;
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

**ETCD 服务发现约定**：
- 注册 key 格式：`/services/transaction-checker/{biz_id}/{instance_id}`
- 注册 value：gRPC 地址，如 `192.168.1.10:8081`
- 租约 TTL：30s，业务方定期续约
- 回查时通过 `biz_id` 前缀查找所有实例，负载均衡选择一个调用
- 如果 ETCD 中找不到对应 `biz_id` 的服务，记录告警日志，retry_count++，下次再查

## 数据模型

### 通知记录表 notifications

```sql
CREATE TABLE notifications (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    key_field VARCHAR(256) NOT NULL DEFAULT '',
    biz_id VARCHAR(64) NOT NULL DEFAULT '',
    channel TINYINT NOT NULL DEFAULT 0,
    -- 通用接收者字段（SMS存手机号，Email存邮箱，站内存用户ID字符串）
    receiver VARCHAR(256) NOT NULL DEFAULT '',
    -- 站内通知冗余 user_id，方便查询和缓存（仅 channel=1 时有值）
    user_id BIGINT NOT NULL DEFAULT 0,
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
    INDEX idx_user_ctime (user_id, ctime DESC),
    INDEX idx_user_group (user_id, group_type, ctime DESC),
    INDEX idx_user_unread (user_id, is_read, ctime DESC),
    INDEX idx_receiver_channel (receiver, channel, ctime DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**关键设计说明**：
- `receiver`：通用接收者字段，存储手机号/邮箱/用户ID字符串，按渠道语义不同。
- `user_id`：站内通知冗余字段（channel=IN_APP 时从 receiver 转换填充），保持与现有 SSE、Redis 缓存、BFF 的 int64 兼容。SMS/Email 渠道此字段为 0。
- `uk_key_channel`：幂等保证。key 为空时由服务端生成 UUID，避免冲突。
- `idx_user_*`：站内通知查询走 user_id 索引。
- `idx_receiver_channel`：SMS/Email 查询走 receiver 索引。

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

### 数据库迁移策略

现有 notifications 表需要迁移。由于是全量重构，采用以下策略：
1. 新建上述三张表（新表名或直接替换）
2. 旧 `notifications` 表中的 `type` 字段映射到 `group_type`：Like(1)/Collect(2) → Interaction(1)，Comment(3)/Reply(4) → Reply(2)，Follow(5) → Follow(4)，Mention(6) → Mention(3)，System(8) → System(5)，Feed(7) → System(5)
3. 旧表 `uid` 字段同时写入 `user_id` 和 `receiver`（receiver 存储 uid 字符串形式）
4. 通过 GORM AutoMigrate 创建新表，旧表数据按需迁移或标记归档

## Domain 模型定义

### Notification

```go
type Notification struct {
    Id             int64
    Key            string
    BizId          string
    Channel        Channel
    Receiver       string              // 通用接收者
    UserId         int64               // 站内通知冗余 user_id
    TemplateId     string
    TemplateParams map[string]string
    Content        string              // 渲染后的内容
    Status         NotificationStatus
    Strategy       SendStrategy
    // 站内通知专用
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
```

### Template

```go
type Template struct {
    Id                    int64
    TemplateId            string          // 业务方使用的模板编码
    Channel               Channel
    Name                  string
    Content               string          // 模板内容，含 {{.变量名}} 占位符
    Description           string
    Status                TemplateStatus
    SMSSign               string          // SMS专用：短信签名
    SMSProviderTemplateId string          // SMS专用：三方平台模板ID
    Ctime                 int64
    Utime                 int64
}
```

### Transaction

```go
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
```

### PrepareRequest

```go
type PrepareRequest struct {
    Notification       Notification
    BizId              string
    CheckBackTimeoutMs int64
}
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
    SendStrategyScheduled SendStrategy = 2  // 预留
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

// SMSProvider 短信服务商接口
// 使用 map[string]string 兼容阿里云（命名参数）和腾讯云（有序参数）
type SMSProvider interface {
    Send(ctx context.Context, tplId string, params map[string]string, numbers ...string) error
}
```

实现类：AliyunSMSProvider（将 map 序列化为 JSON）、TencentSMSProvider（将 map 按 key 排序取 value 列表）。

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
- **InAppSender**：写 DB + 更新 Redis 未读计数缓存 + SSE 推送
- **SMSSender**：调用阿里云/腾讯云 SMS SDK
- **EmailSender**：预留，返回 `errs.ErrNotImplemented`

NotificationService 路由逻辑：

```go
func (s *notificationService) Send(ctx context.Context, n domain.Notification) error {
    // 1. key 为空时生成 UUID
    if n.Key == "" {
        n.Key = uuid.New().String()
    }
    // 2. 站内通知：将 receiver 转为 user_id
    if n.Channel == domain.ChannelInApp {
        uid, err := strconv.ParseInt(n.Receiver, 10, 64)
        if err != nil {
            return fmt.Errorf("站内通知 receiver 必须是合法的用户ID: %w", err)
        }
        n.UserId = uid
    }
    // 3. 幂等检查（key + channel）
    // 4. 模板渲染
    content, err := s.templateSvc.Render(ctx, n.TemplateId, n.Channel, n.TemplateParams)
    if err != nil {
        return err
    }
    n.Content = content
    // 5. 写入 DB，status = Sending
    // 6. 路由到对应渠道发送
    sender := s.senders[n.Channel]
    err = sender.Send(ctx, n)
    // 7. 更新 status = Sent / Failed
    return err
}
```

## TCC 事务流程

### 状态流转

```
普通发送:  Init → Sending → Sent/Failed

TCC 发送:
  Prepare: 写 notifications(status=Init) + transactions(status=Prepared)
           ↓
  Confirm: 更新 transactions(status=Confirmed)
           → 提交 DB 事务
           → 异步发送（模板渲染 + ChannelSender.Send）
           → 更新 notifications(status=Sent/Failed)
           ↓
  Cancel:  更新 transactions(status=Cancelled) + notifications(status=Failed)

回查补偿:
  Prepared 超时 → CheckBack
           ├── Commit   → 走 Confirm 流程
           ├── Rollback → 走 Cancel 流程
           └── Pending  → 延长超时，下次再查
```

### Prepare

```go
func (s *notificationService) Prepare(ctx context.Context, req domain.PrepareRequest) error {
    // 1. key 为空时生成 UUID
    // 2. 幂等检查（key + channel）
    // 3. 在同一个本地 DB 事务中：
    //    a. 写 notifications 表，status = Init
    //    b. 写 notification_transactions 表，status = Prepared
    //       next_check_time = now + check_back_timeout_ms
    return nil
}
```

### Confirm

```go
func (s *notificationService) Confirm(ctx context.Context, key string) error {
    // 1. 查 notification_transactions，校验 status == Prepared
    // 2. 本地 DB 事务：
    //    a. 更新 transactions status = Confirmed
    //    b. 更新 notifications status = Sending
    // 3. DB 事务提交后，再执行外部调用：
    //    a. 模板渲染
    //    b. 路由到 ChannelSender 发送
    //    c. 更新 notifications status = Sent / Failed
    // 注意：外部调用（SMS API 等）不能放在 DB 事务内
    return nil
}
```

### Cancel

```go
func (s *notificationService) Cancel(ctx context.Context, key string) error {
    // 1. 查 notification_transactions，校验 status == Prepared
    // 2. 本地 DB 事务：
    //    a. 更新 transactions status = Cancelled
    //    b. 更新 notifications status = Failed
    return nil
}
```

### Confirm 失败补偿

如果 Confirm 的 DB 事务提交成功，但后续外部发送失败（如 SMS API 超时），notifications 状态为 Sending。定时任务扫描长时间处于 Sending 状态的记录，重新尝试发送。

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
   - 通过 `biz_id` 从 ETCD 查询 key 前缀 `/services/transaction-checker/{biz_id}/`
   - 如果找不到服务实例 → 记录告警日志，retry_count++，下次再查
   - 调用 `CheckTransaction(key)`
   - 根据返回结果：
     - Commit → 调用 Confirm(key)
     - Rollback → 调用 Cancel(key)
     - Pending → retry_count++，next_check_time = now + retryInterval * retry_count（递增退避）

退避策略：基数 10s，第 1 次 10s 后重试，第 2 次 20s，第 3 次 30s...最多 5 次后强制 Cancel。

## 错误处理策略

### gRPC 错误码

| 场景 | gRPC Status Code | 说明 |
|------|------------------|------|
| key 重复（幂等命中） | `ALREADY_EXISTS` | 返回已存在的 notification_id |
| 模板不存在 | `NOT_FOUND` | template_id + channel 未找到 |
| 模板已禁用 | `FAILED_PRECONDITION` | 模板 status=disabled |
| 渠道不支持（Email 预留） | `UNIMPLEMENTED` | Email 渠道未实现 |
| 事务状态不符 | `FAILED_PRECONDITION` | 如 Confirm 时 status 不是 Prepared |
| receiver 格式错误 | `INVALID_ARGUMENT` | 站内通知 receiver 不是合法 int64 |
| key 为空且未自动生成 | `INVALID_ARGUMENT` | 理论上不会出现（服务端兜底生成） |

### SMS 发送失败

- 单次发送失败：更新 notifications status=Failed，不自动重试（由业务方决定是否重发）
- 批量发送部分失败：每条独立记录状态，BatchSendResponse 返回成功的 notification_ids，失败的通过日志记录

### 限流

SMS 渠道在 SMSProvider 实现层面加限流装饰器，复用现有 sms 服务的 `ratelimit` 思路，限制单位时间内的发送频率。

## Kafka 消费者适配

现有 5 个 Kafka 消费者需要适配新的数据模型。改造方式：消费者不再直接操作 Repository，改为调用 `NotificationService.Send`。

映射规则：

| 消费者 | 旧 NotificationType | 新 NotificationGroup | 说明 |
|--------|---------------------|---------------------|------|
| LikeEventConsumer | Like(1) | Interaction(1) | 互动消息 |
| CollectEventConsumer | Collect(2) | Interaction(1) | 互动消息 |
| CommentEventConsumer | Comment(3)/Reply(4) | Reply(2) | 回复我的 |
| FollowEventConsumer | Follow(5) | Follow(4) | 关注 |
| NotificationEventConsumer | System(8) | System(5) | 系统通知 |

消费者改造后的伪代码（以 LikeEventConsumer 为例）：

```go
func (c *LikeEventConsumer) Consume(msg *sarama.ConsumerMessage) error {
    var evt LikeEvent
    json.Unmarshal(msg.Value, &evt)
    // 过滤自己给自己点赞
    if evt.Liked == evt.Liker {
        return nil
    }
    return c.notificationSvc.Send(ctx, domain.Notification{
        Key:       fmt.Sprintf("like:%d:%d", evt.Liker, evt.Liked),
        Channel:   domain.ChannelInApp,
        Receiver:  strconv.FormatInt(evt.Liked, 10),
        TemplateId: "like_notification",  // 需要预置对应模板
        TemplateParams: map[string]string{
            "likerName":   evt.LikerName,
            "targetTitle": evt.TargetTitle,
        },
        GroupType:  domain.NotificationGroupInteraction,
        SourceId:   evt.Liker,
        SourceName: evt.LikerName,
        TargetId:   evt.TargetId,
        TargetType: evt.TargetType,
        TargetTitle: evt.TargetTitle,
    })
}
```

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
│   ├── notification.go        // Notification 领域模型
│   ├── template.go            // Template 领域模型
│   ├── transaction.go         // Transaction 领域模型
│   └── types.go               // 所有枚举定义
├── grpc/
│   └── server.go              // gRPC 服务实现
├── service/
│   ├── notification.go        // NotificationService 核心编排
│   └── template.go            // TemplateService 模板管理+渲染
├── repository/
│   ├── notification.go        // CachedNotificationRepository
│   ├── template.go            // CachedTemplateRepository
│   ├── transaction.go         // TransactionRepository
│   ├── dao/
│   │   ├── init.go            // AutoMigrate 新表
│   │   ├── notification.go
│   │   ├── template.go
│   │   └── transaction.go
│   └── cache/
│       ├── notification.go    // Redis 未读计数（按 group 维度）
│       └── template.go        // Redis 模板缓存
├── channel/
│   ├── types.go               // ChannelSender 接口定义
│   ├── inapp.go               // InAppSender
│   ├── sms.go                 // SMSSender
│   ├── sms_provider.go        // SMSProvider 接口 + 阿里云/腾讯云实现
│   └── email.go               // EmailSender（预留）
├── scheduler/
│   └── checkback.go           // CheckBackScheduler 事务回查
├── events/
│   ├── types.go               // 事件结构体
│   ├── consumer.go            // 5 个 Kafka 消费者（适配新模型）
│   └── producer.go
└── ioc/
    ├── db.go
    ├── redis.go
    ├── kafka.go
    ├── grpc.go
    ├── logger.go
    ├── channel.go             // 渠道 Sender 初始化
    ├── scheduler.go           // 回查任务初始化
    └── template.go            // 模板服务初始化
```

## 保留的现有能力

- Kafka 消费者（like, collect, comment, follow, notification 事件）继续工作，消费者改为调用 NotificationService.Send，适配新的 group_type 和模板参数
- SSE 实时推送（Redis Pub/Sub）保持不变，InAppSender 内部调用现有的 SSE 发布逻辑
- BFF 层 handler 和 SSE Hub 保持不变，Redis 未读计数缓存改为按 NotificationGroup 维度统计
- GetUnreadCount 响应从 `by_type map[NotificationType]count` 改为 `by_group map[NotificationGroup]count`，BFF handler 对应调整
