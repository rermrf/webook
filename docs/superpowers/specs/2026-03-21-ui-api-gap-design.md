# 设计稿接口缺口补全设计文档

## 概述

针对 Pencil 设计稿与后端 BFF 接口的适配分析，补全 5 个缺口：浏览历史模块、IM 会话聚合用户信息、在线状态查询、收藏/点赞列表、标签关注功能。

## 缺口 1：浏览历史模块（新建）

### 背景

设计稿 Browse History 页面显示按日期分组的浏览记录（文章标题、作者、时间、标签），但后端完全没有浏览历史功能。

### 架构

独立微服务 `history/`，MySQL + GORM，gRPC 接口，BFF 提供 REST API。

### 数据模型

```sql
CREATE TABLE browse_histories (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    biz VARCHAR(64) NOT NULL DEFAULT 'article',
    biz_id BIGINT NOT NULL,
    biz_title VARCHAR(256) NOT NULL DEFAULT '',
    author_name VARCHAR(128) NOT NULL DEFAULT '',
    ctime BIGINT NOT NULL DEFAULT 0,
    utime BIGINT NOT NULL DEFAULT 0,
    UNIQUE KEY uk_user_biz (user_id, biz, biz_id),
    INDEX idx_user_ctime (user_id, ctime DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

重复浏览同一文章时 `utime` 更新为当前时间，`ctime` 保持首次浏览时间。通过 `uk_user_biz` 做 upsert。

### gRPC 接口

```protobuf
service HistoryService {
  rpc Record(RecordRequest) returns (RecordResponse);
  rpc List(ListRequest) returns (ListResponse);
  rpc Clear(ClearRequest) returns (ClearResponse);
}

message RecordRequest {
  int64 user_id = 1;
  string biz = 2;
  int64 biz_id = 3;
  string biz_title = 4;
  string author_name = 5;
}
message RecordResponse {}

message ListRequest {
  int64 user_id = 1;
  int64 cursor = 2;
  int32 limit = 3;
}
message ListResponse {
  repeated HistoryItem items = 1;
  bool has_more = 2;
}

message ClearRequest {
  int64 user_id = 1;
}
message ClearResponse {}

message HistoryItem {
  int64 id = 1;
  string biz = 2;
  int64 biz_id = 3;
  string biz_title = 4;
  string author_name = 5;
  int64 ctime = 6;
  int64 utime = 7;
}
```

### BFF REST 接口

```
POST   /history/record                — 记录浏览（文章详情页加载时异步调用）
GET    /history/list?cursor=0&limit=20 — 浏览历史列表（按 utime DESC 游标分页）
DELETE /history                        — 清空浏览历史
```

### 记录时机

在 BFF 的 `ArticleHandler.Detail` 中，查询文章详情成功后，异步（goroutine）调用 `HistoryService.Record`，不阻塞主流程。

### 目录结构

```
history/
├── main.go
├── app.go
├── wire.go
├── wire_gen.go
├── config/
├── domain/
│   └── history.go
├── grpc/
│   └── server.go
├── service/
│   └── history.go
├── repository/
│   ├── history.go
│   └── dao/
│       ├── init.go
│       └── history.go
└── ioc/
    ├── db.go
    ├── grpc.go
    └── logger.go
```

## 缺口 2：IM 会话聚合用户信息

### 背景

设计稿 Messages 页面每个会话显示对方头像和昵称，但 `GET /im/conversations` 返回的 `ConversationVO` 只有 `members []int64`，没有用户信息。

### 方案

修改 `bff/handler/im.go` 的 `ListConversations`，在 BFF 层聚合用户信息。

### 改动

**新增 VO 字段：**

```go
type ConversationVO struct {
    ConversationID string     `json:"conversation_id"`
    Members        []int64    `json:"members"`
    PeerUser       *PeerUser  `json:"peer_user"`
    LastMsg        *MessageVO `json:"last_msg,omitempty"`
    UnreadCount    int64      `json:"unread_count"`
    Utime          int64      `json:"utime"`
}

type PeerUser struct {
    UserId   int64  `json:"user_id"`
    Nickname string `json:"nickname"`
    Avatar   string `json:"avatar"`
}
```

**BFF handler 逻辑：**

1. 调用 `imSvc.ListConversations` 拿到会话列表
2. 从每个会话的 `Members` 中找出对方 userId（排除当前用户）
3. 批量调用 `userSvc.GetProfile` 获取昵称和头像
4. 组装 `PeerUser` 到 `ConversationVO`

**依赖**：`IMHandler` 需要新增 `userSvc userv1.UserServiceClient` 依赖，Wire 中注入。

## 缺口 3：在线状态查询

### 背景

设计稿 Chat Detail 页面显示对方"在线"状态。Redis 已有 `im:online:{userId}` 但没有暴露查询接口。

### 改动

**IM 服务 gRPC 新增：**

```protobuf
rpc IsOnline(IsOnlineRequest) returns (IsOnlineResponse);

message IsOnlineRequest {
  int64 user_id = 1;
}
message IsOnlineResponse {
  bool online = 1;
}
```

IM 服务实现：查询 Redis `im:online:{userId}` 是否存在。

**涉及文件：**
- `api/proto/im/v1/im.proto` — 新增 RPC 和消息
- `im/grpc/server.go` — 实现 IsOnline
- `im/repository/cache/im.go` — 已有 `IsOnline` 方法

**BFF REST 新增：**

```
GET /im/online/:userId  — 返回 {"online": true/false}
```

`bff/handler/im.go` 新增 `GetOnlineStatus` handler。

## 缺口 4：收藏/点赞列表接口

### 背景

设计稿 Profile 页面有收藏和点赞两个 Tab，需要查询"用户收藏过的文章列表"和"用户点赞过的文章列表"。当前互动服务只有单篇查询，缺少列表查询。

### 改动

**互动服务 gRPC 新增：**

```protobuf
rpc ListUserLiked(ListUserLikedRequest) returns (ListUserLikedResponse);
rpc ListUserCollected(ListUserCollectedRequest) returns (ListUserCollectedResponse);

message ListUserLikedRequest {
  int64 user_id = 1;
  string biz = 2;
  int64 offset = 3;
  int64 limit = 4;
}
message ListUserLikedResponse {
  repeated int64 biz_ids = 1;
}

message ListUserCollectedRequest {
  int64 user_id = 1;
  string biz = 2;
  int64 offset = 3;
  int64 limit = 4;
}
message ListUserCollectedResponse {
  repeated int64 biz_ids = 1;
}
```

**涉及文件：**
- `api/proto/intr/v1/interactive.proto` — 新增 RPC
- `interactive/service/` — 新增列表查询方法
- `interactive/repository/dao/` — 新增 DAO 查询（从 user_likes / user_collections 表按 user_id 分页）

**BFF REST 新增：**

```
GET /interactive/liked?offset=0&limit=20     — 我点赞过的文章ID列表
GET /interactive/collected?offset=0&limit=20 — 我收藏过的文章ID列表
```

BFF 层拿到 `biz_ids` 后，批量调用 `ArticleService.GetByIds` 获取文章详情拼装返回。

## 缺口 5：标签关注功能

### 背景

设计稿 Tag Detail 页面显示"关注话题"按钮和关注人数，但 tag 服务没有关注/取关标签的功能。

### 数据模型

```sql
CREATE TABLE tag_follows (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    tag_id BIGINT NOT NULL,
    ctime BIGINT NOT NULL DEFAULT 0,
    UNIQUE KEY uk_user_tag (user_id, tag_id),
    INDEX idx_tag_id (tag_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### gRPC 新增

```protobuf
rpc FollowTag(FollowTagRequest) returns (FollowTagResponse);
rpc UnfollowTag(UnfollowTagRequest) returns (UnfollowTagResponse);
rpc CountTagFollowers(CountTagFollowersRequest) returns (CountTagFollowersResponse);
rpc IsFollowingTag(IsFollowingTagRequest) returns (IsFollowingTagResponse);

message FollowTagRequest {
  int64 user_id = 1;
  int64 tag_id = 2;
}
message FollowTagResponse {}

message UnfollowTagRequest {
  int64 user_id = 1;
  int64 tag_id = 2;
}
message UnfollowTagResponse {}

message CountTagFollowersRequest {
  int64 tag_id = 1;
}
message CountTagFollowersResponse {
  int64 count = 1;
}

message IsFollowingTagRequest {
  int64 user_id = 1;
  int64 tag_id = 2;
}
message IsFollowingTagResponse {
  bool following = 1;
}
```

**涉及文件：**
- `api/proto/tag/v1/tag.proto` — 新增 RPC
- `tag/repository/dao/` — 新增 `tag_follows` DAO
- `tag/service/` — 新增关注逻辑
- `tag/grpc/` — 实现新 RPC

**BFF REST 新增：**

```
POST   /tags/follow    — 关注标签 {"tag_id": 1}
DELETE /tags/follow     — 取消关注 {"tag_id": 1}
```

`bff/handler/tag.go` 的 `GetTagById`（话题详情页）响应中附加 `follower_count` 和 `is_following` 字段。

## 文件变更总结

### 新建文件

| 模块 | 文件 |
|------|------|
| history 微服务 | `history/` 整个目录（domain/service/repository/dao/grpc/ioc/wire/main） |
| history proto | `api/proto/history/v1/history.proto` |
| BFF history handler | `bff/handler/history.go` |
| BFF history ioc | `bff/ioc/history.go` |

### 修改文件

| 文件 | 改动 |
|------|------|
| `api/proto/im/v1/im.proto` | 新增 IsOnline RPC |
| `im/grpc/server.go` | 实现 IsOnline |
| `api/proto/intr/v1/interactive.proto` | 新增 ListUserLiked/ListUserCollected RPC |
| `interactive/service/` | 新增列表查询 |
| `interactive/repository/dao/` | 新增 DAO 列表查询 |
| `api/proto/tag/v1/tag.proto` | 新增 FollowTag/UnfollowTag/CountTagFollowers/IsFollowingTag RPC |
| `tag/repository/dao/` | 新增 tag_follows DAO |
| `tag/service/` | 新增关注逻辑 |
| `tag/grpc/` | 实现新 RPC |
| `bff/handler/im.go` | ListConversations 聚合用户信息 + GetOnlineStatus |
| `bff/handler/tag.go` | 新增关注/取关路由 + 详情页附加关注信息 |
| `bff/handler/article.go` | Detail 中异步记录浏览历史 |
| `bff/ioc/web.go` | 注册 HistoryHandler |
| `bff/wire.go` | 新增 HistorySet |
