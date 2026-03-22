# 图片系统：用户头像、文章封面、图片上传

## 背景

当前系统中用户没有头像字段，文章没有封面图字段。UI 设计稿中 18 个页面几乎所有涉及用户展示的地方都需要头像，文章卡片需要封面图。需要新增图片上传能力和相关数据字段。

## 存储方案

使用 RustFS（S3 兼容对象存储），后续可无缝切换到阿里云/腾讯云 OSS。BFF 层通过 AWS S3 SDK（`github.com/aws/aws-sdk-go-v2`）直接操作 RustFS。

### Bucket 规划

| Bucket | 用途 | Key 格式 | 大小限制 |
|--------|------|---------|---------|
| `avatars` | 用户头像 | `{uid}/avatar.jpg` | 2MB |
| `covers` | 文章封面 | `articles/{article_id}/cover.jpg` | 5MB |
| `images` | 文章内嵌图片 | `articles/{article_id}/{uuid}.jpg` | 10MB |

DB 中只存相对路径 key，完整 URL 在 BFF 层拼接：`{endpoint}/{bucket}/{key}`。

---

## 数据层改动

### 数据库

`users` 表增加：
```sql
ALTER TABLE users ADD COLUMN avatar_url VARCHAR(512) DEFAULT '';
```

`articles` 表增加：
```sql
ALTER TABLE articles ADD COLUMN cover_url VARCHAR(512) DEFAULT '';
```

`published_articles` 表增加：
```sql
ALTER TABLE published_articles ADD COLUMN cover_url VARCHAR(512) DEFAULT '';
```

### Proto 改动

**user.proto** — User message 增加：
```protobuf
string avatar_url = 10;
```

**article.proto** — Article message 增加：
```protobuf
string cover_url = 8;
```

**notification.proto** — Notification message 增加：
```protobuf
string source_avatar_url = 12;
```

**search/sync.proto** — User message 增加：
```protobuf
string avatar_url = 5;
```

---

## BFF 新增接口

| 方法 | 路径 | 认证 | Content-Type | 说明 |
|------|------|------|-------------|------|
| `POST` | `/upload/avatar` | ✅ | multipart/form-data | 上传用户头像，自动更新 users 表 |
| `POST` | `/upload/image` | ✅ | multipart/form-data | 上传通用图片（文章内嵌图） |
| `POST` | `/upload/cover` | ✅ | multipart/form-data | 上传文章封面 |

### 上传流程

1. 前端 multipart/form-data 上传图片（字段名 `file`）
2. BFF 校验：文件类型（jpg/png/webp/gif）、文件大小
3. BFF 上传到 RustFS 对应 bucket
4. 返回图片完整 URL
5. 头像上传额外调用 `userSvc.EditNoSensitive` 更新 `avatar_url`

### 请求/响应

**上传头像：**
```
POST /upload/avatar
Content-Type: multipart/form-data
Body: file=<binary>

Response: { "code": 2, "msg": "OK", "data": { "url": "http://localhost:9000/avatars/123/avatar.jpg" } }
```

**上传图片：**
```
POST /upload/image
Content-Type: multipart/form-data
Body: file=<binary>

Response: { "code": 2, "msg": "OK", "data": { "url": "http://localhost:9000/images/articles/456/uuid.jpg" } }
```

**上传封面：**
```
POST /upload/cover?article_id=456
Content-Type: multipart/form-data
Body: file=<binary>

Response: { "code": 2, "msg": "OK", "data": { "url": "http://localhost:9000/covers/articles/456/cover.jpg" } }
```

---

## BFF 响应补充头像字段

所有涉及用户展示的 VO 增加 `avatar_url`：

| VO 结构体 | 增加字段 |
|----------|---------|
| `ArticleVO` | `CoverUrl string` + `AuthorAvatarUrl string` |
| `Comment` | `UserAvatarUrl string` |
| Profile 响应 | `AvatarUrl string` |
| PublicProfile 响应 | `AvatarUrl string` |
| `RecommendUserVO` | `AvatarUrl string` |
| 通知列表响应 | `SourceAvatarUrl string` |

### enrichment 方式

头像通过 BFF 层 `userSvc.Profile` 获取后填充到响应中，与现有的 authorName enrichment 模式一致。各微服务不需要存储头像信息。

---

## 基础设施

### docker-compose.infra.yaml 新增 RustFS

```yaml
rustfs:
  image: rustfs/rustfs:latest
  ports:
    - "9000:9000"
    - "9001:9001"
  environment:
    RUSTFS_ROOT_USER: admin
    RUSTFS_ROOT_PASSWORD: admin123456
  volumes:
    - rustfs_data:/data
  command: server /data --console-address ":9001"
```

### BFF 配置

```yaml
oss:
  endpoint: "http://localhost:9000"
  access_key: "admin"
  secret_key: "admin123456"
  region: "us-east-1"
  use_path_style: true
```

---

## 涉及文件清单

### Proto + 生成
| 文件 | 改动 |
|------|------|
| `api/proto/user/v1/user.proto` | User 加 avatar_url |
| `api/proto/article/v1/article.proto` | Article 加 cover_url |
| `api/proto/notification/v1/notification.proto` | Notification 加 source_avatar_url |
| `api/proto/search/v1/sync.proto` | User 加 avatar_url |
| `api/proto/gen/*/` | 重新生成 |

### User 微服务
| 文件 | 改动 |
|------|------|
| `user/repository/dao/types.go` | User DAO 结构加 AvatarUrl |
| `user/domain/user.go` | User domain 加 AvatarUrl |
| `user/service/` | Profile/Edit 透传 avatar_url |
| `user/grpc/` | 适配新 proto |

### Article 微服务
| 文件 | 改动 |
|------|------|
| `article/repository/dao/types.go` | Article DAO 加 CoverUrl |
| `article/domain/` | Article domain 加 CoverUrl |
| `article/service/` | Save/Publish 透传 cover_url |
| `article/grpc/` | 适配新 proto |

### BFF 层
| 文件 | 改动 |
|------|------|
| `bff/handler/upload.go` | **新建** — 上传 handler（avatar/image/cover） |
| `bff/ioc/oss.go` | **新建** — S3 客户端初始化 |
| `bff/handler/user.go` | Profile 响应加 AvatarUrl |
| `bff/handler/article.go` | enrichment 加 AuthorAvatarUrl、CoverUrl |
| `bff/handler/article_vo.go` | ArticleVO/Comment 加图片字段 |
| `bff/ioc/web.go` | 注册 upload routes |
| `bff/config/` | 加 OSS 配置 |

### 基础设施
| 文件 | 改动 |
|------|------|
| `docker-compose.infra.yaml` | 新增 RustFS 服务 |
| `script/mysql/init.sql` | users/articles 表加字段 |

---

## 验证方式

1. `go build ./user/...` — user 微服务编译通过
2. `go build ./article/...` — article 微服务编译通过
3. `go build ./bff/...` — BFF 编译通过
4. RustFS 容器启动正常，能通过 S3 API 访问
5. 上传头像接口返回正确 URL
6. Profile 接口返回 avatar_url
7. 文章列表接口返回 cover_url 和 author_avatar_url
