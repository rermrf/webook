# WeBook 后端补全设计文档

## 概述

基于现有后端架构的分析，本文档列出了需要补全和完善的接口，以支持完整的社区功能。

---

## 现有 API 清单

### 用户模块 (User)
| 接口 | 方法 | 路径 | 状态 |
|------|------|------|------|
| 注册 | POST | `/users/signup` | ✅ 已实现 |
| 登录 | POST | `/users/login` | ✅ 已实现 |
| 发送验证码 | POST | `/users/login_sms/code/send` | ✅ 已实现 |
| 短信登录 | POST | `/users/login_sms` | ✅ 已实现 |
| 退出登录 | POST | `/users/logout` | ✅ 已实现 |
| 刷新Token | POST | `/users/refresh_token` | ✅ 已实现 |
| 获取个人资料 | GET | `/users/profile` | ✅ 已实现 |
| 获取他人资料 | GET | `/users/profile/:id` | ✅ 已实现 (新增) |
| 编辑个人资料 | POST | `/users/edit` | ✅ 已实现 |

### 文章模块 (Article)
| 接口 | 方法 | 路径 | 状态 |
|------|------|------|------|
| 保存草稿 | POST | `/articles/edit` | ✅ 已实现 |
| 发布文章 | POST | `/articles/publish` | ✅ 已实现 |
| 撤回文章 | POST | `/articles/withdraw` | ✅ 已实现 |
| 删除文章 | DELETE | `/articles/:id` | ✅ 已实现 (新增) |
| 作者文章列表 | POST | `/articles/list` | ✅ 已实现 |
| 文章详情(作者) | GET | `/articles/detail/:id` | ✅ 已实现 |
| 公开文章列表 | GET | `/articles/pub/articles` | ✅ 已实现 |
| 公开文章详情 | GET | `/articles/pub/:id` | ✅ 已实现 |
| 点赞/取消 | POST | `/articles/pub/like` | ✅ 已实现 |
| 收藏/取消 | POST | `/articles/pub/collect` | ✅ 已实现 |
| 打赏 | POST | `/articles/pub/reward` | ✅ 已实现 |
| 获取互动数据 | GET | `/articles/pub/interactive` | ✅ 已实现 |
| 获取评论 | GET | `/articles/pub/comment` | ✅ 已实现 |
| 获取评论数 | GET | `/articles/pub/comment_cnt` | ✅ 已实现 |
| 发表评论 | POST | `/articles/pub/comment` | ✅ 已实现 |
| 删除评论 | DELETE | `/articles/pub/comment/:id` | ✅ 已实现 (新增) |
| 获取更多回复 | GET | `/articles/pub/comment/:id/replies` | ✅ 已实现 (新增) |

### 关注模块 (Follow)
| 接口 | 方法 | 路径 | 状态 |
|------|------|------|------|
| 关注 | POST | `/follow/follow` | ✅ 已实现 |
| 取消关注 | POST | `/follow/cancel` | ✅ 已实现 |
| 关注列表 | GET | `/follow/followee` | ✅ 已实现 |
| 粉丝列表 | GET | `/follow/follower` | ✅ 已实现 |
| 关注统计 | GET | `/follow/static` | ✅ 已实现 |
| 检查关注状态 | GET | `/follow/check` | ✅ 已实现 (新增) |

### 通知模块 (Notification)
| 接口 | 方法 | 路径 | 状态 |
|------|------|------|------|
| SSE实时推送 | GET | `/notifications/stream` | ✅ 已实现 |
| 通知列表 | GET | `/notifications/list` | ✅ 已实现 |
| 未读通知 | GET | `/notifications/unread` | ✅ 已实现 |
| 未读数统计 | GET | `/notifications/unread-count` | ✅ 已实现 |
| 标记已读 | POST | `/notifications/read/:id` | ✅ 已实现 |
| 全部已读 | POST | `/notifications/read-all` | ✅ 已实现 |
| 删除通知 | DELETE | `/notifications/:id` | ✅ 已实现 |

### 积分模块 (Credit)
| 接口 | 方法 | 路径 | 状态 |
|------|------|------|------|
| 查询余额 | GET | `/credit/balance` | ✅ 已实现 |
| 积分流水 | POST | `/credit/flows` | ✅ 已实现 |
| 每日状态 | GET | `/credit/daily-status` | ✅ 已实现 |
| 发起充值 | POST | `/credit/recharge` | ✅ 已实现 |
| 充值状态 | POST | `/credit/recharge/status` | ✅ 已实现 |
| 积分打赏 | POST | `/credit/reward` | ✅ 已实现 |
| 打赏详情 | POST | `/credit/reward/detail` | ✅ 已实现 |

### 搜索模块 (Search)
| 接口 | 方法 | 路径 | 状态 |
|------|------|------|------|
| 综合搜索 | GET | `/search` | ✅ 已实现 |

### 打赏模块 (Reward)
| 接口 | 方法 | 路径 | 状态 |
|------|------|------|------|
| 查询打赏状态 | POST | `/reward/detail` | ✅ 已实现 |

### 微信登录 (OAuth2)
| 接口 | 方法 | 路径 | 状态 |
|------|------|------|------|
| 获取授权URL | GET | `/oauth2/wechat/authurl` | ✅ 已实现 |
| 回调处理 | ANY | `/oauth2/wechat/callback` | ✅ 已实现 |

### 标签模块 (Tag) - 新增
| 接口 | 方法 | 路径 | 状态 |
|------|------|------|------|
| 获取用户标签 | GET | `/tags` | ✅ 已实现 (新增) |
| 创建标签 | POST | `/tags` | ✅ 已实现 (新增) |
| 绑定标签 | POST | `/tags/attach` | ✅ 已实现 (新增) |
| 获取资源标签 | GET | `/tags/biz` | ✅ 已实现 (新增) |

### Feed 模块 - 新增
| 接口 | 方法 | 路径 | 状态 |
|------|------|------|------|
| 获取 Feed 流 | GET | `/feed` | ✅ 已实现 (新增) |

### 排行榜模块 (Ranking) - 新增
| 接口 | 方法 | 路径 | 状态 |
|------|------|------|------|
| 热门文章榜 | GET | `/ranking/hot` | ✅ 已实现 (新增) |

---

## 已完成的新增接口
```

**BFF Handler 实现**:
```go
// bff/handler/user.go

func (h *UserHandler) ProfileById(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
    idStr := ctx.Param("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        return ginx.Result{Code: 4, Msg: "参数错误"}, err
    }

    resp, err := h.svc.Profile(ctx, &userv1.ProfileRequest{Id: id})
    if err != nil {
        return ginx.Result{Code: 5, Msg: "系统错误"}, err
    }

    user := resp.GetUser()
    // 脱敏处理：隐藏敏感信息
    profile := Profile{
        Id:       user.GetId(),
        Nickname: user.GetNickName(),
        AboutMe:  user.GetAboutMe(),
        Ctime:    user.GetCtime().AsTime().Format(time.DateOnly),
        // 不返回 Email, Phone 等敏感信息
    }
    return ginx.Result{Data: profile}, nil
}
```

**路由注册**:
```go
ug.GET("/profile/:id", ginx.WrapClaims(h.l, h.ProfileById))
```

#### 1.2 批量查询用户资料 (User Service)

**问题**: 关注/粉丝列表接口需要逐个查询用户资料，性能低下。

**Proto 定义**:
```protobuf
// api/proto/user/v1/user.proto

message BatchProfileRequest {
  repeated int64 ids = 1;
}

message BatchProfileResponse {
  map<int64, User> users = 1;
}

service UserService {
  // 新增
  rpc BatchProfile(BatchProfileRequest) returns (BatchProfileResponse);
}
```

#### 1.3 修改密码

**建议新增**:
```
POST /users/password
```

**Proto 定义**:
```protobuf
message ChangePasswordRequest {
  int64 uid = 1;
  string old_password = 2;
  string new_password = 3;
}

message ChangePasswordResponse {
}

service UserService {
  rpc ChangePassword(ChangePasswordRequest) returns (ChangePasswordResponse);
}
```

#### 1.4 头像上传

**建议新增**:
```
POST /users/avatar
```

需要集成对象存储服务 (OSS/MinIO)。

---

### 2. 文章模块扩展

#### 2.1 删除文章

**建议新增**:
```
DELETE /articles/:id
```

**Proto 定义**:
```protobuf
message DeleteRequest {
  int64 id = 1;
  int64 author_id = 2; // 权限校验
}

message DeleteResponse {
}

service ArticleService {
  rpc Delete(DeleteRequest) returns (DeleteResponse);
}
```

#### 2.2 文章浏览历史

**建议新增**:
```
GET /articles/history
POST /articles/history/record  (阅读时自动记录)
DELETE /articles/history/:id
DELETE /articles/history/clear
```

需要新增 History 模块或扩展现有 history 模块。

#### 2.3 收藏夹管理

**建议新增**:
```
GET /collections                   # 收藏夹列表
POST /collections                  # 创建收藏夹
PUT /collections/:id               # 更新收藏夹
DELETE /collections/:id            # 删除收藏夹
GET /collections/:id/articles      # 收藏夹内文章
```

**Proto 定义** (扩展 interactive 模块):
```protobuf
message Collection {
  int64 id = 1;
  int64 uid = 2;
  string name = 3;
  string description = 4;
  bool is_public = 5;
  int64 article_count = 6;
  google.protobuf.Timestamp ctime = 7;
  google.protobuf.Timestamp utime = 8;
}

service InteractiveService {
  // 新增
  rpc CreateCollection(CreateCollectionRequest) returns (CreateCollectionResponse);
  rpc ListCollections(ListCollectionsRequest) returns (ListCollectionsResponse);
  rpc GetCollectionArticles(GetCollectionArticlesRequest) returns (GetCollectionArticlesResponse);
  // 修改 Collect 接口，支持 collection_id
}
```

#### 2.4 热门/推荐文章

**建议新增**:
```
GET /articles/pub/hot        # 热门文章
GET /articles/pub/recommend  # 推荐文章
```

可利用现有 ranking 模块:
```go
// bff/handler/article.go

func (h *ArticleHandler) HotArticles(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
    resp, err := h.rankingSvc.TopN(ctx, &rankingv1.TopNRequest{})
    if err != nil {
        return ginx.Result{Code: 5, Msg: "系统错误"}, err
    }
    // 转换返回
}
```

#### 2.5 文章标签

**建议新增**:
```
GET /tags                    # 获取用户标签
POST /tags                   # 创建标签
POST /articles/:id/tags      # 为文章添加标签
DELETE /articles/:id/tags    # 移除文章标签
GET /articles/by-tag/:tagId  # 按标签查询文章
```

已有 tag 模块的 proto 定义，需要在 BFF 层暴露接口:

```go
// bff/handler/tag.go

type TagHandler struct {
    svc tagv1.TagServiceClient
    l   logger.LoggerV1
}

func (h *TagHandler) RegisterRoutes(server *gin.Engine) {
    g := server.Group("/tags")
    g.GET("", ginx.WrapClaims(h.l, h.GetUserTags))
    g.POST("", ginx.WrapBodyAndToken(h.l, h.CreateTag))
    g.POST("/attach", ginx.WrapBodyAndToken(h.l, h.AttachTags))
    g.GET("/biz", ginx.WrapBody(h.l, h.GetBizTags))
}
```

---

### 3. 评论模块扩展

#### 3.1 删除评论 (BFF 层)

**建议新增**:
```
DELETE /articles/pub/comment/:id
```

```go
// bff/handler/article.go

func (h *ArticleHandler) DeleteComment(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
    idStr := ctx.Param("id")
    id, _ := strconv.ParseInt(idStr, 10, 64)

    _, err := h.commentSvc.DeleteComment(ctx, &commentv1.DeleteCommentRequest{
        Id: id,
        // 需要在 comment 服务层校验是否为评论作者
    })
    if err != nil {
        return ginx.Result{Code: 5, Msg: "系统错误"}, err
    }
    return ginx.Result{Msg: "OK"}, nil
}
```

#### 3.2 获取更多回复

**建议新增**:
```
GET /articles/pub/comment/:id/replies
```

已有 proto 定义 `GetMoreReplies`，需要在 BFF 暴露:

```go
func (h *ArticleHandler) GetMoreReplies(ctx *gin.Context, req GetMoreRepliesReq) (ginx.Result, error) {
    ridStr := ctx.Param("id")
    rid, _ := strconv.ParseInt(ridStr, 10, 64)

    resp, err := h.commentSvc.GetMoreReplies(ctx, &commentv1.GetMoreRepliesRequest{
        Rid:   rid,
        MinId: req.MinId,
        Limit: req.Limit,
    })
    if err != nil {
        return ginx.Result{Code: 5, Msg: "系统错误"}, err
    }
    // 转换返回
}
```

#### 3.3 评论点赞

**建议新增**:
```
POST /articles/pub/comment/:id/like
```

需要扩展 comment 或 interactive 模块支持评论点赞。

---

### 4. 关注模块扩展

#### 4.1 检查关注状态

**建议新增**:
```
GET /follow/check?followee=xxx
```

```go
// bff/handler/follow.go

func (h *FollowHandler) CheckFollow(ctx *gin.Context, req FolloweeRequest, uc ijwt.UserClaims) (ginx.Result, error) {
    resp, err := h.svc.FollowInfo(ctx, &followv1.FollowInfoRequest{
        Follower: uc.UserId,
        Followee: req.Followee,
    })
    if err != nil {
        // 可能是未关注，返回 false
        return ginx.Result{Data: map[string]bool{"followed": false}}, nil
    }
    return ginx.Result{
        Data: map[string]bool{"followed": resp.FollowRelation != nil},
    }, nil
}
```

---

### 5. Feed 模块

#### 5.1 个人 Feed 流

**建议新增**:
```
GET /feed
```

已有 feed proto 定义，需要在 BFF 暴露:

```go
// bff/handler/feed.go

type FeedHandler struct {
    svc feedv1.FeedSvcClient
    l   logger.LoggerV1
}

func (h *FeedHandler) RegisterRoutes(server *gin.Engine) {
    server.GET("/feed", ginx.WrapBodyAndToken(h.l, h.GetFeed))
}

func (h *FeedHandler) GetFeed(ctx *gin.Context, req GetFeedReq, uc ijwt.UserClaims) (ginx.Result, error) {
    resp, err := h.svc.FindFeedEvents(ctx, &feedv1.FindFeedEventsRequest{
        Uid:       uc.UserId,
        Limit:     req.Limit,
        Timestamp: req.Timestamp,
    })
    if err != nil {
        return ginx.Result{Code: 5, Msg: "系统错误"}, err
    }
    // 转换返回
}
```

---

### 6. 消息/私信模块 (新增)

#### 6.1 Proto 定义

```protobuf
// api/proto/message/v1/message.proto
syntax = "proto3";

package message.v1;
option go_package = "/message/v1;messagev1";

import "google/protobuf/timestamp.proto";

service MessageService {
  // 发送私信
  rpc Send(SendRequest) returns (SendResponse);
  // 获取会话列表
  rpc ListConversations(ListConversationsRequest) returns (ListConversationsResponse);
  // 获取消息历史
  rpc ListMessages(ListMessagesRequest) returns (ListMessagesResponse);
  // 标记已读
  rpc MarkAsRead(MarkAsReadRequest) returns (MarkAsReadResponse);
  // 获取未读数
  rpc GetUnreadCount(GetUnreadCountRequest) returns (GetUnreadCountResponse);
}

message Message {
  int64 id = 1;
  int64 sender_id = 2;
  int64 receiver_id = 3;
  string content = 4;
  bool is_read = 5;
  google.protobuf.Timestamp ctime = 6;
}

message Conversation {
  int64 id = 1;
  int64 user_id = 2;  // 对方用户ID
  string user_name = 3;
  string last_message = 4;
  int64 unread_count = 5;
  google.protobuf.Timestamp last_time = 6;
}

// ... 请求/响应消息定义
```

#### 6.2 BFF 接口

```
POST /messages/send               # 发送私信
GET /messages/conversations       # 会话列表
GET /messages/:userId             # 与某人的消息记录
POST /messages/:userId/read       # 标记与某人的消息已读
GET /messages/unread-count        # 未读私信数
```

---

### 7. 举报模块 (新增)

#### 7.1 Proto 定义

```protobuf
// api/proto/report/v1/report.proto
syntax = "proto3";

package report.v1;
option go_package = "/report/v1;reportv1";

service ReportService {
  rpc Create(CreateRequest) returns (CreateResponse);
  rpc List(ListRequest) returns (ListResponse);  // 管理员接口
  rpc Handle(HandleRequest) returns (HandleResponse);  // 管理员接口
}

message Report {
  int64 id = 1;
  int64 reporter_id = 2;  // 举报人
  string target_type = 3; // article, comment, user
  int64 target_id = 4;
  string reason = 5;      // 举报原因分类
  string description = 6; // 详细描述
  int32 status = 7;       // 0=待处理, 1=已处理, 2=已驳回
  int64 ctime = 8;
}
```

#### 7.2 BFF 接口

```
POST /reports             # 提交举报
GET /reports              # 我的举报记录
```

---

### 8. 管理后台接口 (新增)

#### 8.1 用户管理
```
GET /admin/users              # 用户列表
GET /admin/users/:id          # 用户详情
PUT /admin/users/:id/status   # 禁用/启用用户
```

#### 8.2 文章审核
```
GET /admin/articles/pending    # 待审核文章
PUT /admin/articles/:id/review # 审核文章
DELETE /admin/articles/:id     # 删除文章
```

#### 8.3 评论管理
```
GET /admin/comments            # 评论列表
DELETE /admin/comments/:id     # 删除评论
```

#### 8.4 数据统计
```
GET /admin/stats/overview      # 概览数据
GET /admin/stats/users         # 用户增长
GET /admin/stats/articles      # 文章统计
```

---

## 已有模块完善建议

### 1. Notification 服务

**问题**: Kafka 消费者需要正确处理通知事件。

**建议**:
- 确保 `bff/events/notification_producer.go` 正确发送事件
- 在 notification 服务中实现 Kafka 消费者
- 支持通知聚合（如：多人点赞同一文章只显示一条）

### 2. Credit 服务

**问题**: 积分获取逻辑 (`EarnCredit`) 需要与各业务模块集成。

**建议**:
- 在文章阅读时触发积分获取
- 在点赞、收藏、评论时触发积分获取
- 使用 Kafka 事件异步处理，避免阻塞主流程

```go
// 示例：在文章阅读时发送积分事件
go func() {
    producer.ProduceEarnCreditEvent(ctx, EarnCreditEvent{
        Uid:   uc.UserId,
        Biz:   "read",
        BizId: articleId,
    })
}()
```

### 3. OpenAPI 服务

**问题**: 开放平台功能已有 proto 定义，但缺少 BFF 层接口。

**建议完善**:
```
# 应用管理
GET /openapi/apps             # 我的应用列表
POST /openapi/apps            # 创建应用
GET /openapi/apps/:id         # 应用详情
PUT /openapi/apps/:id         # 更新应用
DELETE /openapi/apps/:id      # 删除应用
POST /openapi/apps/:id/secret # 重置密钥

# 授权管理
GET /openapi/authorizations   # 已授权应用列表
DELETE /openapi/authorizations/:appId # 取消授权
```

---

## 安全性建议

### 1. 接口权限校验

确保所有需要登录的接口都正确校验了用户身份:

```go
// middleware 层面
func AuthRequired() gin.HandlerFunc {
    return func(c *gin.Context) {
        claims, ok := c.Get("claims")
        if !ok {
            c.AbortWithStatusJSON(401, ginx.Result{Code: 4, Msg: "未登录"})
            return
        }
        // 继续
        c.Next()
    }
}
```

### 2. 资源所有权校验

修改/删除操作需要校验资源所有者:

```go
// 示例：删除文章时校验作者
func (h *ArticleHandler) Delete(ctx *gin.Context, req DeleteReq, uc ijwt.UserClaims) (ginx.Result, error) {
    artResp, err := h.svc.GetById(ctx, &articlev1.GetByIdRequest{Id: req.Id})
    if err != nil {
        return ginx.Result{Code: 5, Msg: "系统错误"}, err
    }
    if artResp.GetArticle().GetAuthor().GetId() != uc.UserId {
        return ginx.Result{Code: 4, Msg: "无权限"}, nil
    }
    // 执行删除
}
```

### 3. 频率限制

对敏感接口添加频率限制:

```go
// 发送验证码
ug.POST("/login_sms/code/send",
    middleware.RateLimit(1, time.Minute), // 每分钟1次
    ginx.WrapBody(h.l, h.SendLoginSMSCode))

// 发表评论
pub.POST("/comment",
    middleware.RateLimit(10, time.Minute), // 每分钟10次
    ginx.WrapBodyAndToken(h.l, h.CreateComment))
```

### 4. 敏感内容过滤

评论、文章内容需要过滤敏感词:

```go
// pkg/sensitive/filter.go
type Filter interface {
    Check(content string) (bool, []string) // 返回是否包含敏感词及敏感词列表
    Replace(content string) string          // 替换敏感词
}

// 在创建评论前调用
if hasSensitive, words := filter.Check(req.Content); hasSensitive {
    return ginx.Result{Code: 4, Msg: "内容包含敏感词"}, nil
}
```

---

## Wire 依赖注入更新

需要在 `bff/wire.go` 中添加新的 handler 注入:

```go
// bff/wire.go

var ProviderSet = wire.NewSet(
    // 现有 providers...

    // 新增
    NewTagHandler,
    NewFeedHandler,
    // NewMessageHandler,
    // NewReportHandler,
)

func InitApp() *gin.Engine {
    wire.Build(
        ProviderSet,
        // ...
    )
    return nil
}
```

---

## 数据库迁移

### 新增表

```sql
-- 私信表
CREATE TABLE messages (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    sender_id BIGINT NOT NULL,
    receiver_id BIGINT NOT NULL,
    content TEXT NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    ctime DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_sender_receiver (sender_id, receiver_id),
    INDEX idx_receiver_read (receiver_id, is_read)
);

-- 举报表
CREATE TABLE reports (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    reporter_id BIGINT NOT NULL,
    target_type VARCHAR(32) NOT NULL,
    target_id BIGINT NOT NULL,
    reason VARCHAR(64) NOT NULL,
    description TEXT,
    status TINYINT DEFAULT 0,
    ctime DATETIME DEFAULT CURRENT_TIMESTAMP,
    utime DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_status (status),
    INDEX idx_target (target_type, target_id)
);

-- 收藏夹表
CREATE TABLE collections (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    uid BIGINT NOT NULL,
    name VARCHAR(64) NOT NULL,
    description VARCHAR(256),
    is_public BOOLEAN DEFAULT TRUE,
    article_count INT DEFAULT 0,
    ctime DATETIME DEFAULT CURRENT_TIMESTAMP,
    utime DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_uid (uid)
);

-- 收藏夹-文章关联表
CREATE TABLE collection_articles (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    collection_id BIGINT NOT NULL,
    article_id BIGINT NOT NULL,
    ctime DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_collection_article (collection_id, article_id)
);
```

---

## 实现优先级

### P0 (必须)
1. 获取他人资料接口
2. 批量查询用户资料
3. 文章删除
4. 评论删除
5. 检查关注状态
6. Feed 接口暴露
7. Tag 接口暴露

### P1 (重要)
1. 文章浏览历史
2. 收藏夹管理
3. 热门/推荐文章
4. 评论回复加载
5. 修改密码
6. 头像上传

### P2 (可选)
1. 私信功能
2. 举报功能
3. 评论点赞
4. 管理后台接口
5. OpenAPI 完善

---

## 测试建议

1. **单元测试**: 每个新增接口需要覆盖核心逻辑
2. **集成测试**: 验证 BFF 层与 gRPC 服务的交互
3. **压力测试**: Feed 流、通知推送等高频接口
4. **安全测试**: 权限校验、SQL 注入、XSS 等
