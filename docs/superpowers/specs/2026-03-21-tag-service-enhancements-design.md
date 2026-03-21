# Tag 服务完善：关注话题、批量查询、排序优化、权限校验

## 背景

Tag 服务已从用户个人标签改造为文章全局标签。核心功能（创建标签、绑定文章、按标签查文章）已就位。审查发现 5 处与 UI 设计稿不匹配的缺口，需要补齐。

## 改动总览

| # | 功能 | 性质 | 优先级 |
|---|------|------|--------|
| 1 | 标签关注（关注/取消/检查/我的关注/关注数） | 新增 | 高 |
| 2 | 批量获取文章标签 | 新增 | 高 |
| 3 | "最热"排序完善 | 改动 | 中 |
| 4 | "精选"排序实现 | 改动 | 中 |
| 5 | AttachTags 权限校验 | 改动 | 中 |

---

## 1. 标签关注功能

### 数据库

`webook_tag` 库新增 `tag_follows` 表：

```sql
CREATE TABLE tag_follows (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  uid BIGINT NOT NULL,
  tag_id BIGINT NOT NULL,
  ctime BIGINT DEFAULT 0,
  UNIQUE KEY uk_uid_tag (uid, tag_id),
  INDEX idx_tag_id (tag_id)
);
```

`tags` 表增加字段：

```sql
ALTER TABLE tags ADD COLUMN follower_count BIGINT DEFAULT 0;
```

### Proto

```protobuf
rpc FollowTag(FollowTagRequest) returns (FollowTagResponse);
rpc UnfollowTag(UnfollowTagRequest) returns (UnfollowTagResponse);
rpc CheckTagFollow(CheckTagFollowRequest) returns (CheckTagFollowResponse);
rpc GetUserFollowedTags(GetUserFollowedTagsRequest) returns (GetUserFollowedTagsResponse);

message FollowTagRequest {
  int64 uid = 1;
  int64 tag_id = 2;
}
message FollowTagResponse {}

message UnfollowTagRequest {
  int64 uid = 1;
  int64 tag_id = 2;
}
message UnfollowTagResponse {}

message CheckTagFollowRequest {
  int64 uid = 1;
  int64 tag_id = 2;
}
message CheckTagFollowResponse {
  bool followed = 1;
}

message GetUserFollowedTagsRequest {
  int64 uid = 1;
  int32 offset = 2;
  int32 limit = 3;
}
message GetUserFollowedTagsResponse {
  repeated Tag tags = 1;
}
```

Tag message 增加 `follower_count`：

```protobuf
message Tag {
  int64 id = 1;
  string name = 2;
  string description = 3;
  int64 follower_count = 4;  // 新增
}
```

### BFF 接口

| 方法 | 路径 | 认证 | 请求 | 响应 |
|------|------|------|------|------|
| `POST` | `/tags/:id/follow` | ✅ | 无 | `{success: true}` |
| `DELETE` | `/tags/:id/follow` | ✅ | 无 | `{success: true}` |
| `GET` | `/tags/:id/follow` | ✅ | 无 | `{followed: bool}` |
| `GET` | `/tags/followed` | ✅ | `?offset=0&limit=20` | `[{id, name, description, follower_count}]` |

### 实现逻辑

- **关注**：事务内 `tag_follows` INSERT + `tags.follower_count` +1。重复关注幂等（INSERT IGNORE）。
- **取消关注**：事务内 `tag_follows` DELETE + `tags.follower_count` -1。未关注时幂等（不报错）。
- **检查**：查询 `tag_follows` 表 `WHERE uid=? AND tag_id=?` 是否存在。
- **我的关注**：`SELECT t.* FROM tags t JOIN tag_follows tf ON t.id = tf.tag_id WHERE tf.uid=? ORDER BY tf.ctime DESC LIMIT ? OFFSET ?`。
- **GetTagById** 返回值自动包含 `follower_count`（从 tags 表读取）。

### Domain

```go
type Tag struct {
    Id            int64
    Name          string
    Description   string
    FollowerCount int64  // 新增
}
```

### DAO 新增

```go
type TagFollow struct {
    Id    int64 `gorm:"primaryKey,autoIncrement"`
    Uid   int64 `gorm:"uniqueIndex:uk_uid_tag"`
    TagId int64 `gorm:"uniqueIndex:uk_uid_tag;index:idx_tag_id"`
    Ctime int64
}
```

DAO 接口新增 4 个方法：

```go
FollowTag(ctx context.Context, uid, tagId int64) error
UnfollowTag(ctx context.Context, uid, tagId int64) error
CheckTagFollow(ctx context.Context, uid, tagId int64) (bool, error)
GetUserFollowedTags(ctx context.Context, uid int64, offset, limit int) ([]Tag, error)
```

---

## 2. 批量获取文章标签

### Proto

```protobuf
rpc BatchGetBizTags(BatchGetBizTagsRequest) returns (BatchGetBizTagsResponse);

message BatchGetBizTagsRequest {
  string biz = 1;
  repeated int64 biz_ids = 2;
}

message BizTagList {
  repeated Tag tags = 1;
}

message BatchGetBizTagsResponse {
  map<int64, BizTagList> biz_tags = 1;
}
```

### BFF 接口

| 方法 | 路径 | 认证 | 请求 | 响应 |
|------|------|------|------|------|
| `POST` | `/tags/batch_biz` | ❌ | `{biz: "article", biz_ids: [1,2,3]}` | `{biz_tags: {1: [{id,name}], 2: [...]}}` |

### DAO 实现

```go
func (d *GORMTagDao) BatchGetTagsByBiz(ctx context.Context, biz string, bizIds []int64) (map[int64][]Tag, error)
```

SQL：`SELECT tb.biz_id, t.* FROM tag_bizs tb JOIN tags t ON tb.tid = t.id WHERE tb.biz = ? AND tb.biz_id IN (?) ORDER BY tb.biz_id`

在代码中按 `biz_id` 分组为 `map[int64][]Tag`。

### 限制

- `biz_ids` 最多 100 个，超出返回错误。

---

## 3. "最热"排序完善

### 现状

`GetBizIdsByTag` 的 `hottest` 排序 `ORDER BY biz_id DESC`，不是真正的热度排序。

### 改为

在 **tag service 层**（非 DAO 层）实现：

1. 从 DAO 获取该标签下所有 biz_id（或最近 1000 条）
2. 调用 `intrSvc.GetByIds()` 批量获取互动数据
3. 内存计算热度分数：`score = like_cnt * 3 + collect_cnt * 5 + read_cnt * 0.1`
4. 排序后取 offset/limit 分页返回

### 依赖

tag service 需要注入 `intrv1.InteractiveServiceClient`：

```go
type tagService struct {
    repo    repository.TagRepository
    producer events.Producer
    intrSvc  intrv1.InteractiveServiceClient  // 新增
    l        logger.LoggerV1
}
```

### 注意

- 数据量大时（>1000 篇文章），先取最近 1000 条再排序。
- 热度排序结果不缓存（实时性要求）。

---

## 4. "精选"排序实现

### 公式

精选 = 质量 + 时间衰减：

```
hours = (now - ctime) / 3600
score = (like_cnt * 3 + collect_cnt * 5) / (hours + 2)^1.5
```

### 实现

与"最热"共享代码路径。`GetBizIdsByTag` 的 `sortBy` 参数增加 `featured` 分支：

```go
switch sortBy {
case "newest":
    // DAO 直接 ORDER BY ctime DESC
case "hottest":
    // service 层计算热度
case "featured":
    // service 层计算精选分数（需要 ctime + 互动数据）
}
```

需要同时获取文章的 `ctime`，调用 `articleSvc.GetPublishedById` 或在 `tag_bizs` 表中记录 ctime。

**推荐方案：** 在 `tag_bizs` 表增加 `biz_ctime` 字段，`AttachTags` 时传入，避免精选排序需要额外调用 article 服务。

```sql
ALTER TABLE tag_bizs ADD COLUMN biz_ctime BIGINT DEFAULT 0;
```

---

## 5. AttachTags 权限校验

### BFF handler 层改动

```go
func (h *TagHandler) AttachTags(ctx *gin.Context, req AttachTagsReq, uc ijwt.UserClaims) {
    if req.Biz == "article" {
        // 调用 articleSvc.GetById() 验证文章作者
        artResp, err := h.articleSvc.GetById(ctx, &articlev1.GetByIdRequest{Id: req.BizId})
        if err != nil {
            // 返回 500
        }
        if artResp.GetArticle().GetAuthor().GetId() != uc.UserId {
            // 返回 403：只有作者可以设置标签
        }
    }
    // 继续调用 tagSvc.AttachTags
}
```

### 依赖

TagHandler 需要注入 `articlev1.ArticleServiceClient`：

```go
type TagHandler struct {
    tagSvc     tagv1.TagServiceClient
    articleSvc articlev1.ArticleServiceClient  // 新增
    l          logger.LoggerV1
}
```

---

## 涉及文件清单

### Tag 微服务
| 文件 | 改动 |
|------|------|
| `api/proto/tag/v1/tag.proto` | Tag 加 follower_count；新增 4 个关注 RPC + BatchGetBizTags RPC |
| `tag/domain/tag.go` | Tag 增加 FollowerCount |
| `tag/repository/dao/types.go` | 新增 TagFollow 结构体；Tag 增加 FollowerCount；TagBiz 增加 BizCtime；DAO 接口新增 5 个方法 |
| `tag/repository/dao/gorm.go` | 实现 FollowTag/UnfollowTag/CheckTagFollow/GetUserFollowedTags/BatchGetTagsByBiz |
| `tag/repository/types.go` | Repository 接口新增 5 个方法 |
| `tag/repository/tag.go` | 实现新增方法 |
| `tag/service/types.go` | Service 接口新增 5 个方法 |
| `tag/service/tag.go` | 实现关注逻辑 + 热度/精选排序（注入 intrSvc） |
| `tag/grpc/tag.go` | 适配新 proto，实现 5 个新 RPC |
| `tag/ioc/` | 更新依赖注入，注入 intrSvc |

### BFF 层
| 文件 | 改动 |
|------|------|
| `bff/handler/tag.go` | 新增 4 个关注路由 + 1 个批量查询路由 + AttachTags 权限校验 |
| `bff/ioc/web.go` | TagHandler 构造增加 articleSvc 参数 |

### Proto 重新生成
| 文件 | 改动 |
|------|------|
| `api/proto/gen/tag/v1/tag.pb.go` | 重新生成 |
| `api/proto/gen/tag/v1/tag_grpc.pb.go` | 重新生成 |

## 验证方式

1. `go build ./tag/...` — tag 微服务编译通过
2. `go build ./bff/...` — BFF 编译通过
3. 检查新增 DAO 的 SQL 正确性
4. 验证关注/取消关注的幂等性
5. 验证热度排序和精选排序返回正确顺序
