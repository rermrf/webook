# IM 私信模块设计文档（第一期）

## 概述

自研 IM 私信模块，替换原有的 OpenIM 集成方案。第一期实现 1v1 私信功能（文本 + 图片），后续迭代扩展在线状态和群聊。

## 背景

### 现有状态

- `im/` 目录仅有 OpenIM 用户同步代码（REST 调用 + Kafka binlog 监听），无消息处理能力
- 项目中无 WebSocket 实现
- 无 MongoDB 依赖（将新引入）

### 设计目标

1. 1v1 私信：文本消息和图片消息的发送与接收
2. 会话管理：会话列表、最后一条消息摘要、未读计数
3. 实时推送：WebSocket 长连接 + Redis Pub/Sub 跨实例转发
4. 消息历史：游标分页查询历史消息
5. 消息撤回：支持撤回已发送的消息
6. 多设备支持：同一用户多个设备同时在线

### 技术选型

| 组件 | 选型 | 说明 |
|------|------|------|
| 消息存储 | MongoDB | 文档模型适合消息存储，支持灵活查询 |
| 缓存/Pub/Sub | Redis | 未读计数 + 会话排序 + 跨实例消息转发 |
| 实时推送 | WebSocket (gorilla/websocket) | 全双工通信 |
| 跨实例转发 | Redis Pub/Sub | 多 BFF 实例间消息同步 |
| 服务间通信 | gRPC | 与项目现有架构一致 |

## 整体架构

```
前端 (WebSocket)
    │
BFF 层 (WebSocket Hub + REST API)
    │
    ├── WebSocket Hub ────→ Redis Pub/Sub ────→ 其他 BFF 实例的 Hub
    │   (管理连接, 转发消息)
    │
    ├── REST API (会话列表/历史消息/未读数)
    │       │
    │       ▼
    │   IM gRPC Service
    │       │
    │       ├── MessageService       (消息收发)
    │       ├── ConversationService  (会话管理)
    │       └── Repository
    │               ├── MongoDB (消息存储 + 会话存储)
    │               └── Redis   (会话缓存/未读计数)
    │
    └── Kafka (预留，异步事件扩展)
```

### 消息发送流程

```
用户A 发送消息
  → BFF WebSocket Hub 接收
  → 调用 IM gRPC Service.SendMessage
    → 写入 MongoDB 消息记录
    → 更新会话 last_msg
    → 更新 Redis 未读计数
    → 返回消息 ID + 时间戳
  → 向发送方推送 ack
  → Redis Publish 到 im:msg:{conversationId}
  → 接收方的 Hub 实例通过 Subscribe 收到
  → 推送给接收方的 WebSocket 连接
```

## 数据模型

### MongoDB 集合

#### messages 集合

```go
type Message struct {
    ID             primitive.ObjectID `bson:"_id,omitempty"`
    ConversationID string             `bson:"conversation_id"`
    SenderId       int64              `bson:"sender_id"`
    ReceiverId     int64              `bson:"receiver_id"`
    MsgType        uint8              `bson:"msg_type"`     // 1=文本 2=图片
    Content        string             `bson:"content"`      // 文本内容或图片URL
    Status         uint8              `bson:"status"`       // 1=已发送 2=已送达 3=已读 4=已撤回
    Ctime          int64              `bson:"ctime"`
}
```

索引：
- `{conversation_id: 1, ctime: -1}` — 按会话查历史消息
- `{sender_id: 1, ctime: -1}` — 按发送者查
- `{receiver_id: 1, status: 1}` — 未送达消息查询

#### conversations 集合

```go
type Conversation struct {
    ID             primitive.ObjectID `bson:"_id,omitempty"`
    ConversationID string             `bson:"conversation_id"` // 格式: "conv:{minUid}:{maxUid}"
    Members        []int64            `bson:"members"`         // [userA, userB]
    LastMsg        LastMessage         `bson:"last_msg"`
    Ctime          int64              `bson:"ctime"`
    Utime          int64              `bson:"utime"`
}

type LastMessage struct {
    Content  string `bson:"content"`
    MsgType  uint8  `bson:"msg_type"`
    SenderId int64  `bson:"sender_id"`
    Ctime    int64  `bson:"ctime"`
}
```

索引：
- `{conversation_id: 1}` — 唯一索引
- `{members: 1, utime: -1}` — 按用户查会话列表

**ConversationID 生成规则**：`conv:{min(userA, userB)}:{max(userA, userB)}`，确保两个用户之间只有一个会话。

### Redis 结构

| Key | 类型 | 说明 |
|-----|------|------|
| `im:unread:{userId}` | Hash | field=conversationId, value=未读数 |
| `im:conv:{userId}` | ZSet | score=最后消息时间, member=conversationId |
| `im:online:{userId}` | String | 值=连接所在实例ID, TTL 30s |

Redis Pub/Sub 频道：`im:msg:{conversationId}` — 实时消息推送

### 枚举定义

```go
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
```

## gRPC API 设计

### 服务接口

```protobuf
service IMService {
  // 消息
  rpc SendMessage(SendMessageRequest) returns (SendMessageResponse);
  rpc ListMessages(ListMessagesRequest) returns (ListMessagesResponse);
  rpc MarkAsRead(MarkAsReadRequest) returns (MarkAsReadResponse);
  rpc RecallMessage(RecallMessageRequest) returns (RecallMessageResponse);

  // 会话
  rpc ListConversations(ListConversationsRequest) returns (ListConversationsResponse);
  rpc GetConversation(GetConversationRequest) returns (GetConversationResponse);
  rpc GetUnreadCount(GetUnreadCountRequest) returns (GetUnreadCountResponse);
}
```

### 请求/响应消息

```protobuf
// ===== 消息 =====
message SendMessageRequest {
  int64 sender_id = 1;
  int64 receiver_id = 2;
  uint32 msg_type = 3;       // 1=文本 2=图片
  string content = 4;
}
message SendMessageResponse {
  string message_id = 1;
  string conversation_id = 2;
  int64 ctime = 3;
}

message ListMessagesRequest {
  string conversation_id = 1;
  int64 cursor = 2;           // 游标（ctime），0 表示从最新开始
  int32 limit = 3;
}
message ListMessagesResponse {
  repeated MessageItem messages = 1;
  bool has_more = 2;
}

message MarkAsReadRequest {
  int64 user_id = 1;
  string conversation_id = 2;
}
message MarkAsReadResponse {}

message RecallMessageRequest {
  int64 user_id = 1;
  string message_id = 2;
}
message RecallMessageResponse {}

// ===== 会话 =====
message ListConversationsRequest {
  int64 user_id = 1;
  int64 cursor = 2;           // 游标（utime），0 表示从最新开始
  int32 limit = 3;
}
message ListConversationsResponse {
  repeated ConversationItem conversations = 1;
  bool has_more = 2;
}

message GetConversationRequest {
  string conversation_id = 1;
}
message GetConversationResponse {
  ConversationItem conversation = 1;
}

message GetUnreadCountRequest {
  int64 user_id = 1;
}
message GetUnreadCountResponse {
  int64 total = 1;
  map<string, int64> by_conversation = 2; // conversationId → 未读数
}

// ===== 通用消息 =====
message MessageItem {
  string id = 1;
  string conversation_id = 2;
  int64 sender_id = 3;
  int64 receiver_id = 4;
  uint32 msg_type = 5;
  string content = 6;
  uint32 status = 7;
  int64 ctime = 8;
}

message ConversationItem {
  string conversation_id = 1;
  repeated int64 members = 2;
  MessageItem last_msg = 3;
  int64 unread_count = 4;
  int64 utime = 5;
}
```

## WebSocket 协议

### 消息格式（JSON）

**客户端 → 服务端：**

```json
// 发送消息
{"action": "send", "conversation_id": "conv:1:2", "msg_type": 1, "content": "你好"}

// 确认收到
{"action": "ack", "message_id": "xxx"}

// 正在输入
{"action": "typing", "conversation_id": "conv:1:2"}

// 心跳
{"action": "heartbeat"}
```

**服务端 → 客户端：**

```json
// 新消息
{"action": "message", "conversation_id": "conv:1:2", "message": {"id": "xxx", "sender_id": 1, "msg_type": 1, "content": "你好", "ctime": 1711000000000}}

// 撤回通知
{"action": "recall", "conversation_id": "conv:1:2", "message_id": "xxx"}

// 发送确认
{"action": "ack", "message_id": "xxx", "ctime": 1711000000000}

// 错误
{"action": "error", "code": 400, "msg": "消息内容不能为空"}
```

### 连接生命周期

```
连接建立:
  → JWT 认证（query param token）
  → 创建 IMClient，注册到 Hub
  → 启动 readPump + writePump goroutine
  → 设置 Redis im:online:{userId}（TTL 30s）

心跳保活:
  → 客户端每 25s 发送 heartbeat
  → 服务端续约 im:online:{userId} TTL
  → 服务端每 30s 发送 WebSocket ping frame

连接断开:
  → readPump 检测到错误或 close
  → 从 Hub 注销
  → 删除 Redis im:online:{userId}（无其他设备连接时）
```

## BFF 层设计

### WebSocket Hub

```go
type IMHub struct {
    clients    map[int64]map[*IMClient]bool  // userId → 多设备连接
    register   chan *IMClient
    unregister chan *IMClient
    redis      *redis.Client
    imSvc      imv1.IMServiceClient
    l          logger.LoggerV1
}

type IMClient struct {
    UserId int64
    Conn   *websocket.Conn
    Send   chan []byte
}
```

**多实例支持**：通过 Redis Pub/Sub，用户 A 在 BFF-1，用户 B 在 BFF-2 时，消息通过 Redis 频道中转。

### REST 接口

```
GET  /im/conversations                   — 会话列表（游标分页）
GET  /im/conversations/:id/messages      — 历史消息（游标分页）
POST /im/conversations/:id/read          — 标记会话已读
GET  /im/unread-count                    — 总未读数
GET  /im/ws                              — WebSocket 连接端点
```

## 目录结构

### IM 微服务

```
im/
├── main.go
├── app.go
├── wire.go
├── wire_gen.go
├── config/
│   ├── dev.yaml
│   └── docker.yaml
├── domain/
│   ├── message.go
│   ├── conversation.go
│   └── types.go
├── grpc/
│   └── server.go
├── service/
│   ├── message.go
│   └── conversation.go
├── repository/
│   ├── message.go
│   ├── conversation.go
│   ├── dao/
│   │   ├── init.go
│   │   ├── message.go
│   │   └── conversation.go
│   └── cache/
│       └── im.go
└── ioc/
    ├── mongo.go
    ├── redis.go
    ├── grpc.go
    └── logger.go
```

### BFF 层新增

```
bff/
├── handler/
│   ├── im.go
│   └── ws/
│       ├── hub.go
│       ├── client.go
│       └── protocol.go
├── ioc/
│   └── im.go
```

### Proto 定义

```
api/proto/
├── im/
│   └── v1/
│       └── im.proto
```

## 删除的旧代码

原有的 OpenIM 集成代码将被完全移除：
- `im/domain/user.go`
- `im/service/user.go`
- `im/events/mysql_binlog_event.go`

## 第二期扩展预留

第一期不实现，但架构上预留：
- **在线状态**：Redis `im:online:{userId}` 已设计，第二期添加状态查询 API 和 UI 展示
- **群聊**：ConversationID 规则可扩展为 `group:{groupId}`，Members 支持多人，第二期添加群管理 API
- **图片消息**：第一期支持，content 存储 OSS URL，上传走独立的文件服务接口
