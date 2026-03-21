# IM 私信模块实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 自研 IM 私信模块，支持 1v1 文本/图片消息收发、会话管理、未读计数，使用 WebSocket 实时推送。

**Architecture:** IM 作为独立 gRPC 微服务（MongoDB + Redis），BFF 层提供 WebSocket Hub（Redis Pub/Sub 跨实例）+ REST API。消息发送通过 WebSocket → BFF Hub → gRPC → MongoDB 持久化 → Redis Pub/Sub → 接收方 WebSocket。

**Tech Stack:** Go, MongoDB (go.mongodb.org/mongo-driver), Redis, WebSocket (gorilla/websocket), gRPC/protobuf, Wire DI, ETCD

**Spec:** `docs/superpowers/specs/2026-03-21-im-private-messaging-design.md`

---

## File Map

### 新建文件

| 文件路径 | 职责 |
|---------|------|
| `im/domain/message.go` | Message 领域模型 |
| `im/domain/conversation.go` | Conversation 领域模型 |
| `im/domain/types.go` | MsgType, MsgStatus 枚举 |
| `im/repository/dao/init.go` | MongoDB 集合初始化 + 索引创建 |
| `im/repository/dao/message.go` | MongoDB message CRUD |
| `im/repository/dao/conversation.go` | MongoDB conversation CRUD |
| `im/repository/cache/im.go` | Redis 未读计数 + 会话排序 |
| `im/repository/message.go` | MessageRepository |
| `im/repository/conversation.go` | ConversationRepository |
| `im/service/message.go` | MessageService |
| `im/service/conversation.go` | ConversationService |
| `im/grpc/server.go` | gRPC 服务实现（7 个 RPC） |
| `im/ioc/mongo.go` | MongoDB 客户端初始化 |
| `im/ioc/redis.go` | Redis 初始化 |
| `im/ioc/grpc.go` | gRPC 服务器初始化 |
| `im/ioc/logger.go` | Logger 初始化 |
| `im/wire.go` | Wire 依赖注入 |
| `im/app.go` | App 结构体 |
| `im/main.go` | 入口 |
| `im/config/dev.yaml` | 开发配置 |
| `im/config/docker.yaml` | Docker 配置 |
| `api/proto/im/v1/im.proto` | Proto 定义 |
| `bff/handler/ws/protocol.go` | WebSocket 消息协议 |
| `bff/handler/ws/client.go` | IMClient（readPump/writePump） |
| `bff/handler/ws/hub.go` | IMHub（连接管理 + Redis Pub/Sub） |
| `bff/handler/im.go` | IM REST handler |
| `bff/ioc/im.go` | IM gRPC 客户端 + Hub 初始化 |

### 删除文件

| 文件路径 | 说明 |
|---------|------|
| `im/domain/user.go` | 旧 OpenIM 用户模型 |
| `im/service/user.go` | 旧 OpenIM REST 同步 |
| `im/events/mysql_binlog_event.go` | 旧 binlog 消费者 |

### 修改文件

| 文件路径 | 改动说明 |
|---------|---------|
| `bff/ioc/web.go` | 注册 IMHandler 路由 |
| `bff/wire.go` | 新增 IMSet Wire 依赖集 |
| `bff/wire_gen.go` | Wire 重新生成 |
| `bff/app.go` | 无改动（Hub 在 IoC 中启动） |

---

## Chunk 1: Domain 层 + Proto + 删除旧代码

### Task 1: 删除旧 OpenIM 代码 + 创建 Domain 层

**Files:**
- Delete: `im/domain/user.go`, `im/service/user.go`, `im/events/mysql_binlog_event.go`
- Create: `im/domain/types.go`, `im/domain/message.go`, `im/domain/conversation.go`

- [ ] **Step 1: 删除旧代码**

删除 `im/domain/user.go`、`im/service/user.go`、`im/events/mysql_binlog_event.go`。如果 `im/events/` 和 `im/service/` 目录为空则一并删除。

- [ ] **Step 2: 创建 `im/domain/types.go`**

```go
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
```

- [ ] **Step 3: 创建 `im/domain/message.go`**

```go
package domain

type Message struct {
	Id             string
	ConversationID string
	SenderId       int64
	ReceiverId     int64
	MsgType        MsgType
	Content        string
	Status         MsgStatus
	Ctime          int64
}
```

- [ ] **Step 4: 创建 `im/domain/conversation.go`**

```go
package domain

import "fmt"

type Conversation struct {
	Id             string
	ConversationID string
	Members        []int64
	LastMsg        LastMessage
	Ctime          int64
	Utime          int64
}

type LastMessage struct {
	Content  string
	MsgType  MsgType
	SenderId int64
	Ctime    int64
}

// GenConversationID 生成会话ID: conv:{min}:{max}
func GenConversationID(userA, userB int64) string {
	if userA > userB {
		userA, userB = userB, userA
	}
	return fmt.Sprintf("conv:%d:%d", userA, userB)
}
```

- [ ] **Step 5: 验证编译**

Run: `cd im && go build ./domain/...`

- [ ] **Step 6: 提交**

```bash
git add im/domain/ && git rm im/service/user.go im/events/mysql_binlog_event.go
git commit -m "feat(im): 删除旧 OpenIM 代码，添加 IM domain 模型"
```

---

### Task 2: 创建 Proto 定义并生成代码

**Files:**
- Create: `api/proto/im/v1/im.proto`

- [ ] **Step 1: 创建 `api/proto/im/v1/im.proto`**

完整 proto 定义，包含：
- `IMService` 服务（7 个 RPC: SendMessage, ListMessages, MarkAsRead, RecallMessage, ListConversations, GetConversation, GetUnreadCount）
- 所有 Request/Response 消息
- `MessageItem` 和 `ConversationItem` 消息

package: `im.v1`
go_package: `/im/v1;imv1`

- [ ] **Step 2: 生成代码**

Run: `buf generate ./api/proto`

验证 `api/proto/gen/im/v1/` 下生成了 `im.pb.go` 和 `im_grpc.pb.go`。

- [ ] **Step 3: 提交**

```bash
git add api/proto/im/ api/proto/gen/im/
git commit -m "feat(im): 添加 IM proto v1 定义和生成代码"
```

---

## Chunk 2: DAO 层 + Cache 层 + Repository 层

### Task 3: MongoDB DAO 层

**Files:**
- Create: `im/repository/dao/init.go`, `im/repository/dao/message.go`, `im/repository/dao/conversation.go`

- [ ] **Step 1: 创建 `im/repository/dao/init.go`**

```go
package dao

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitCollections(db *mongo.Database) error {
	ctx := context.Background()

	// messages 索引
	msgCol := db.Collection("messages")
	_, err := msgCol.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "conversation_id", Value: 1}, {Key: "ctime", Value: -1}}},
		{Keys: bson.D{{Key: "sender_id", Value: 1}, {Key: "ctime", Value: -1}}},
		{Keys: bson.D{{Key: "receiver_id", Value: 1}, {Key: "status", Value: 1}}},
	})
	if err != nil {
		return err
	}

	// conversations 索引
	convCol := db.Collection("conversations")
	_, err = convCol.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "conversation_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "members", Value: 1}, {Key: "utime", Value: -1}}},
	})
	return err
}
```

- [ ] **Step 2: 创建 `im/repository/dao/message.go`**

Message DAO entity（bson 结构体）+ MessageDAO 接口：

```go
type MessageDAO interface {
	Insert(ctx context.Context, msg Message) (string, error)
	FindByConversation(ctx context.Context, conversationId string, cursor int64, limit int) ([]Message, error)
	FindById(ctx context.Context, id string) (Message, error)
	UpdateStatus(ctx context.Context, id string, status uint8) error
	UpdateStatusBatch(ctx context.Context, conversationId string, receiverId int64, fromStatus, toStatus uint8) error
}
```

实现：
- `Insert`: InsertOne，返回 ObjectID hex string
- `FindByConversation`: 按 conversation_id + ctime < cursor 游标分页（cursor=0 时不加条件），ORDER BY ctime DESC，LIMIT
- `FindById`: 按 _id 查询
- `UpdateStatus`: 按 _id 更新 status
- `UpdateStatusBatch`: 按 conversation_id + receiver_id + status 批量更新（用于 MarkAsRead）

- [ ] **Step 3: 创建 `im/repository/dao/conversation.go`**

Conversation DAO entity + ConversationDAO 接口：

```go
type ConversationDAO interface {
	Upsert(ctx context.Context, conv Conversation) error
	FindByConversationId(ctx context.Context, conversationId string) (Conversation, error)
	FindByUserId(ctx context.Context, userId int64, cursor int64, limit int) ([]Conversation, error)
	UpdateLastMsg(ctx context.Context, conversationId string, lastMsg LastMessage) error
}
```

实现：
- `Upsert`: 按 conversation_id 做 upsert（不存在则创建，存在则跳过）
- `FindByConversationId`: 按 conversation_id 查单条
- `FindByUserId`: `members` 数组包含 userId，按 utime DESC 游标分页
- `UpdateLastMsg`: 更新 last_msg + utime

- [ ] **Step 4: 验证编译**

Run: `cd im && go build ./repository/dao/...`

- [ ] **Step 5: 提交**

```bash
git add im/repository/dao/
git commit -m "feat(im): 添加 MongoDB DAO 层（message + conversation）"
```

---

### Task 4: Redis Cache 层

**Files:**
- Create: `im/repository/cache/im.go`

- [ ] **Step 1: 创建 `im/repository/cache/im.go`**

```go
type IMCache interface {
	// 未读计数
	IncrUnread(ctx context.Context, userId int64, conversationId string) error
	ClearUnread(ctx context.Context, userId int64, conversationId string) error
	GetUnread(ctx context.Context, userId int64) (map[string]int64, error)

	// 会话排序
	UpdateConvScore(ctx context.Context, userId int64, conversationId string, score int64) error

	// 在线状态
	SetOnline(ctx context.Context, userId int64, instanceId string) error
	SetOffline(ctx context.Context, userId int64) error
	IsOnline(ctx context.Context, userId int64) (bool, error)

	// Pub/Sub
	Publish(ctx context.Context, conversationId string, data []byte) error
}
```

Redis key 规则：
- `im:unread:{userId}` — Hash（field=conversationId, value=count）
- `im:conv:{userId}` — ZSet（score=lastMsgTime, member=conversationId）
- `im:online:{userId}` — String（TTL 30s）

- [ ] **Step 2: 验证编译**

Run: `cd im && go build ./repository/cache/...`

- [ ] **Step 3: 提交**

```bash
git add im/repository/cache/
git commit -m "feat(im): 添加 IM Redis 缓存层"
```

---

### Task 5: Repository 层

**Files:**
- Create: `im/repository/message.go`, `im/repository/conversation.go`

- [ ] **Step 1: 创建 `im/repository/message.go`**

```go
type MessageRepository interface {
	Create(ctx context.Context, msg domain.Message) (string, error)
	FindByConversation(ctx context.Context, conversationId string, cursor int64, limit int) ([]domain.Message, error)
	FindById(ctx context.Context, id string) (domain.Message, error)
	UpdateStatus(ctx context.Context, id string, status domain.MsgStatus) error
	MarkConversationRead(ctx context.Context, conversationId string, receiverId int64) error
}
```

封装 DAO，做 domain ↔ dao 转换。`MarkConversationRead` 调用 DAO 的 `UpdateStatusBatch`（将 Sent/Delivered 改为 Read）。

- [ ] **Step 2: 创建 `im/repository/conversation.go`**

```go
type ConversationRepository interface {
	CreateIfNotExist(ctx context.Context, conv domain.Conversation) error
	FindByConversationId(ctx context.Context, conversationId string) (domain.Conversation, error)
	FindByUserId(ctx context.Context, userId int64, cursor int64, limit int) ([]domain.Conversation, error)
	UpdateLastMsg(ctx context.Context, conversationId string, lastMsg domain.LastMessage) error
	GetUnreadCount(ctx context.Context, userId int64) (map[string]int64, error)
	IncrUnread(ctx context.Context, userId int64, conversationId string) error
	ClearUnread(ctx context.Context, userId int64, conversationId string) error
	UpdateConvScore(ctx context.Context, userId int64, conversationId string, score int64) error
}
```

封装 DAO + Cache。`GetUnreadCount` 从 Redis Hash 读取。`IncrUnread` 更新 Redis。

- [ ] **Step 3: 验证编译**

Run: `cd im && go build ./repository/...`

- [ ] **Step 4: 提交**

```bash
git add im/repository/message.go im/repository/conversation.go
git commit -m "feat(im): 添加 IM Repository 层"
```

---

## Chunk 3: Service 层 + gRPC Server

### Task 6: Service 层

**Files:**
- Create: `im/service/message.go`, `im/service/conversation.go`

- [ ] **Step 1: 创建 `im/service/message.go`**

```go
type MessageService interface {
	SendMessage(ctx context.Context, msg domain.Message) (domain.Message, error)
	ListMessages(ctx context.Context, conversationId string, cursor int64, limit int) ([]domain.Message, bool, error)
	RecallMessage(ctx context.Context, userId int64, messageId string) error
	MarkAsRead(ctx context.Context, userId int64, conversationId string) error
}
```

`SendMessage` 流程：
1. 生成 ConversationID（`GenConversationID`）
2. 设置 Status=Sent, Ctime=now
3. 调用 ConversationRepo.CreateIfNotExist（确保会话存在）
4. 调用 MessageRepo.Create
5. 更新会话 LastMsg（ConversationRepo.UpdateLastMsg）
6. 对接收方递增未读数（ConversationRepo.IncrUnread）
7. 更新两个用户的会话排序分数（ConversationRepo.UpdateConvScore）
8. 返回完整 Message（含 Id、ConversationID、Ctime）

`MarkAsRead` 流程：
1. MessageRepo.MarkConversationRead
2. ConversationRepo.ClearUnread

- [ ] **Step 2: 创建 `im/service/conversation.go`**

```go
type ConversationService interface {
	ListConversations(ctx context.Context, userId int64, cursor int64, limit int) ([]domain.Conversation, bool, error)
	GetConversation(ctx context.Context, conversationId string) (domain.Conversation, error)
	GetUnreadCount(ctx context.Context, userId int64) (int64, map[string]int64, error)
}
```

`GetUnreadCount` 返回 total + 按会话的未读 map。

`ListConversations` 从 Repository 获取会话列表，同时附加每个会话的未读数。

- [ ] **Step 3: 验证编译**

Run: `cd im && go build ./service/...`

- [ ] **Step 4: 提交**

```bash
git add im/service/
git commit -m "feat(im): 添加 IM Service 层（MessageService + ConversationService）"
```

---

### Task 7: gRPC Server

**Files:**
- Create: `im/grpc/server.go`

- [ ] **Step 1: 创建 gRPC server**

```go
type IMServiceServer struct {
	imv1.UnimplementedIMServiceServer
	msgSvc  service.MessageService
	convSvc service.ConversationService
}

func NewIMServiceServer(msgSvc service.MessageService, convSvc service.ConversationService) *IMServiceServer
func (s *IMServiceServer) Register(server *grpc.Server)
```

实现 7 个 RPC 方法：
- `SendMessage`：调用 msgSvc.SendMessage，返回 message_id + conversation_id + ctime
- `ListMessages`：调用 msgSvc.ListMessages，转换为 MessageItem 列表 + has_more
- `MarkAsRead`：调用 msgSvc.MarkAsRead
- `RecallMessage`：调用 msgSvc.RecallMessage
- `ListConversations`：调用 convSvc.ListConversations，转换为 ConversationItem 列表
- `GetConversation`：调用 convSvc.GetConversation
- `GetUnreadCount`：调用 convSvc.GetUnreadCount

domain ↔ proto 转换函数：`toMessageItem`, `toConversationItem`

- [ ] **Step 2: 验证编译**

Run: `cd im && go build ./grpc/...`

- [ ] **Step 3: 提交**

```bash
git add im/grpc/
git commit -m "feat(im): 添加 IM gRPC Server（7 个 RPC 方法）"
```

---

## Chunk 4: IoC + Wire + 启动

### Task 8: IoC 初始化 + Wire + 启动文件

**Files:**
- Create: `im/ioc/mongo.go`, `im/ioc/redis.go`, `im/ioc/grpc.go`, `im/ioc/logger.go`
- Create: `im/wire.go`, `im/app.go`, `im/main.go`
- Create: `im/config/dev.yaml`, `im/config/docker.yaml`

- [ ] **Step 1: 创建 `im/ioc/mongo.go`**

```go
func InitMongo() *mongo.Database {
	type Config struct {
		URI    string `yaml:"uri"`
		DBName string `yaml:"dbName"`
	}
	// viper.UnmarshalKey("mongo", &cfg)
	// mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI))
	// client.Database(cfg.DBName)
	// dao.InitCollections(db) — 创建索引
}
```

- [ ] **Step 2: 创建 `im/ioc/redis.go`**

复用 notification 的模式：`InitRedis() redis.Cmdable`

- [ ] **Step 3: 创建 `im/ioc/grpc.go`**

```go
func InitGRPCServer(imServer *grpc2.IMServiceServer, l logger.LoggerV1) *grpcx.Server {
	// Name: "im"
	// 读取 grpc.server 配置
}
```

- [ ] **Step 4: 创建 `im/ioc/logger.go`**

复用 notification 的模式：`InitLogger() logger.LoggerV1`

- [ ] **Step 5: 创建 Wire 配置**

```go
// im/wire.go
var thirdPartySet = wire.NewSet(InitMongo, InitLogger, InitRedis)
var daoSet = wire.NewSet(dao.NewMessageDAO, dao.NewConversationDAO)
var cacheSet = wire.NewSet(cache.NewIMCache)
var repoSet = wire.NewSet(repository.NewMessageRepository, repository.NewConversationRepository)
var serviceSet = wire.NewSet(service.NewMessageService, service.NewConversationService)
var serverSet = wire.NewSet(grpc2.NewIMServiceServer, ioc.InitGRPCServer)

func InitApp() *App {
	wire.Build(thirdPartySet, daoSet, cacheSet, repoSet, serviceSet, serverSet, wire.Struct(new(App), "*"))
	return new(App)
}
```

- [ ] **Step 6: 创建 `im/app.go`**

```go
type App struct {
	server *grpcx.Server
}
```

- [ ] **Step 7: 创建 `im/main.go`**

```go
func main() {
	initViper()
	app := InitApp()
	err := app.server.Serve()
	if err != nil { panic(err) }
}
```

- [ ] **Step 8: 创建配置文件**

`im/config/docker.yaml`:
```yaml
mongo:
  uri: "mongodb://mongo:27017"
  dbName: "webook_im"
redis:
  addr: "redis:6379"
grpc:
  server:
    port: 8101
    etcdAddrs:
      - "etcd:2379"
```

- [ ] **Step 9: 安装依赖**

Run: `go get github.com/gorilla/websocket`（WebSocket 库，Chunk 5 需要）

- [ ] **Step 10: 运行 Wire 生成**

Run: `cd im && wire`

- [ ] **Step 11: 验证完整编译**

Run: `cd im && go build .`

- [ ] **Step 12: 提交**

```bash
git add im/
git commit -m "feat(im): 添加 IoC/Wire/启动文件，IM 微服务可编译"
```

---

## Chunk 5: BFF WebSocket Hub + REST API

### Task 9: WebSocket 协议 + Client

**Files:**
- Create: `bff/handler/ws/protocol.go`, `bff/handler/ws/client.go`

- [ ] **Step 1: 创建 `bff/handler/ws/protocol.go`**

WebSocket 消息协议定义：

```go
package ws

// 客户端 → 服务端
type ClientMessage struct {
	Action         string `json:"action"`          // send | ack | typing | heartbeat
	ConversationID string `json:"conversation_id,omitempty"`
	MsgType        uint8  `json:"msg_type,omitempty"`
	Content        string `json:"content,omitempty"`
	MessageID      string `json:"message_id,omitempty"`
}

// 服务端 → 客户端
type ServerMessage struct {
	Action         string       `json:"action"`          // message | recall | ack | error
	ConversationID string       `json:"conversation_id,omitempty"`
	Message        *MessageData `json:"message,omitempty"`
	MessageID      string       `json:"message_id,omitempty"`
	Ctime          int64        `json:"ctime,omitempty"`
	Code           int          `json:"code,omitempty"`
	Msg            string       `json:"msg,omitempty"`
}

type MessageData struct {
	Id         string `json:"id"`
	SenderId   int64  `json:"sender_id"`
	ReceiverId int64  `json:"receiver_id"`
	MsgType    uint8  `json:"msg_type"`
	Content    string `json:"content"`
	Ctime      int64  `json:"ctime"`
}
```

- [ ] **Step 2: 创建 `bff/handler/ws/client.go`**

```go
type IMClient struct {
	UserId int64
	Conn   *websocket.Conn
	Send   chan []byte
	hub    *IMHub
}
```

`readPump`：
- 设置读超时、MaxMessageSize、PongHandler
- 循环读取消息 → JSON 解析 ClientMessage → 按 action 分发
- `send`: 调用 hub.handleSend
- `ack`: 处理确认
- `typing`: 通过 Redis Pub/Sub 转发
- `heartbeat`: 续约在线状态
- 连接断开时从 hub 注销

`writePump`：
- 循环从 Send channel 读取数据 → 写入 WebSocket
- 处理 ticker ping frame（30s）

- [ ] **Step 3: 验证编译**

Run: `cd bff && go build ./handler/ws/...`

- [ ] **Step 4: 提交**

```bash
git add bff/handler/ws/
git commit -m "feat(bff/im): 添加 WebSocket 协议定义和 Client"
```

---

### Task 10: WebSocket Hub

**Files:**
- Create: `bff/handler/ws/hub.go`

- [ ] **Step 1: 创建 IMHub**

```go
type IMHub struct {
	clients    map[int64]map[*IMClient]bool
	register   chan *IMClient
	unregister chan *IMClient
	mu         sync.RWMutex
	redis      *redis.Client
	imSvc      imv1.IMServiceClient
	cache      IMHubCache  // 封装在线状态操作
	l          logger.LoggerV1
}

func NewIMHub(redisClient redis.Cmdable, imSvc imv1.IMServiceClient, l logger.LoggerV1) *IMHub
```

`Run(ctx)` 循环：
- `go h.subscribeRedis(ctx)` — 订阅 `im:msg:*` 频道
- select register/unregister/ctx.Done

`handleSend(client, msg)` 方法：
1. 调用 imSvc.SendMessage gRPC
2. 向发送方推送 ack（包含 message_id + ctime）
3. Redis Publish 到 `im:msg:{conversationId}`（携带完整消息数据和接收者 userId）

`subscribeRedis(ctx)` 方法：
- `PSubscribe("im:msg:*")`
- 收到消息后解析目标 userId
- 查找本地 clients map，推送到对应 IMClient.Send

`Register/Unregister` 方法：注册/注销客户端，管理在线状态

- [ ] **Step 2: 验证编译**

Run: `cd bff && go build ./handler/ws/...`

- [ ] **Step 3: 提交**

```bash
git add bff/handler/ws/hub.go
git commit -m "feat(bff/im): 添加 WebSocket Hub（连接管理 + Redis Pub/Sub）"
```

---

### Task 11: IM REST Handler + IoC

**Files:**
- Create: `bff/handler/im.go`, `bff/ioc/im.go`
- Modify: `bff/ioc/web.go`, `bff/wire.go`

- [ ] **Step 1: 创建 `bff/handler/im.go`**

```go
type IMHandler struct {
	svc imv1.IMServiceClient
	hub *ws.IMHub
	l   logger.LoggerV1
}

func NewIMHandler(svc imv1.IMServiceClient, hub *ws.IMHub, l logger.LoggerV1) *IMHandler

func (h *IMHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/im")
	g.GET("/ws", h.WebSocket)                              // WebSocket 连接
	g.GET("/conversations", ginx.WrapClaims(h.l, h.ListConversations))
	g.GET("/conversations/:id/messages", ginx.WrapClaims(h.l, h.ListMessages))
	g.POST("/conversations/:id/read", ginx.WrapClaims(h.l, h.MarkAsRead))
	g.GET("/unread-count", ginx.WrapClaims(h.l, h.GetUnreadCount))
}
```

`WebSocket` handler：
1. 从 query param 获取 token，验证 JWT
2. `websocket.Upgrader` 升级连接
3. 创建 IMClient，注册到 Hub
4. 启动 readPump + writePump goroutine

REST handlers 调用 gRPC：
- `ListConversations`：调用 imSvc.ListConversations，返回 ConversationVO 列表
- `ListMessages`：调用 imSvc.ListMessages，返回 MessageVO 列表
- `MarkAsRead`：调用 imSvc.MarkAsRead
- `GetUnreadCount`：调用 imSvc.GetUnreadCount

- [ ] **Step 2: 创建 `bff/ioc/im.go`**

```go
func InitIMGRPCClient(client *etcdv3.Client) imv1.IMServiceClient {
	// ETCD 服务发现，连接 "etcd:///service/im"
}

func InitIMHub(redisClient redis.Cmdable, imSvc imv1.IMServiceClient, l logger.LoggerV1) *ws.IMHub {
	hub := ws.NewIMHub(redisClient, imSvc, l)
	go hub.Run(context.Background())
	return hub
}
```

- [ ] **Step 3: 更新 `bff/ioc/web.go`**

在 `InitGin` 函数参数中新增 `imHdl *handler.IMHandler`，并调用 `imHdl.RegisterRoutes(server)`。

- [ ] **Step 4: 更新 `bff/wire.go`**

新增 Wire 依赖集：

```go
var IMSet = wire.NewSet(
	handler.NewIMHandler,
	ioc.InitIMGRPCClient,
	ioc.InitIMHub,
)
```

在 `InitApp()` 的 `wire.Build()` 中加入 `IMSet`。

- [ ] **Step 5: 运行 Wire 生成**

Run: `cd bff && wire`

- [ ] **Step 6: 验证完整编译**

Run: `cd bff && go build .`

- [ ] **Step 7: 提交**

```bash
git add bff/handler/im.go bff/handler/ws/ bff/ioc/im.go bff/ioc/web.go bff/wire.go bff/wire_gen.go
git commit -m "feat(bff/im): 添加 IM REST Handler + WebSocket 集成 + Wire 更新"
```

---

## 依赖关系

```
Task 1 (Domain + 删除旧代码)
  → Task 2 (Proto v1)
    → Task 3 (DAO 层)
      → Task 4 (Cache 层)
        → Task 5 (Repository 层)
          → Task 6 (Service 层)
            → Task 7 (gRPC Server)
              → Task 8 (IoC + Wire + 启动)
                → Task 9 (WS Protocol + Client)
                  → Task 10 (WS Hub)
                    → Task 11 (REST Handler + IoC + Wire)
```
