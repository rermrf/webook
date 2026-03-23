# Image System Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add user avatars, article cover images, and image upload capability using RustFS (S3-compatible storage).

**Architecture:** Proto fields added to user/article messages, DAO/domain layers updated, BFF gets a new upload handler using AWS S3 SDK v1 (already in go.mod), all existing VOs enriched with avatar/cover URLs during BFF enrichment.

**Tech Stack:** Go, gRPC/protobuf, GORM, AWS S3 SDK v1, Gin multipart upload, RustFS, Docker

**Spec:** `docs/superpowers/specs/2026-03-22-image-system-design.md`

---

## File Structure

### New Files
| File | Responsibility |
|------|---------------|
| `bff/handler/upload.go` | Image upload HTTP handlers (avatar, cover, image) |
| `bff/ioc/oss.go` | S3 client initialization for RustFS |

### Modified Files
| File | Responsibility |
|------|---------------|
| `api/proto/user/v1/user.proto` | Add avatar_url to User message |
| `api/proto/article/v1/article.proto` | Add cover_url to Article message |
| `api/proto/notification/v2/notification.proto` | Add source_avatar_url to NotificationItem |
| `user/domain/user.go` | Add AvatarUrl field |
| `user/repository/dao/user.go` | Add AvatarUrl to DAO struct |
| `user/grpc/server.go` | Update toDTO/toV conversions |
| `article/domain/article.go` | Add CoverUrl field |
| `article/repository/dao/gorm.go` | Add CoverUrl to DAO structs |
| `article/grpc/server.go` | Update toDTO/toV conversions |
| `bff/handler/article_vo.go` | Add CoverUrl, AuthorAvatarUrl, UserAvatarUrl |
| `bff/handler/user.go` | Add AvatarUrl to Profile/PublicProfile/RecommendUserVO |
| `bff/handler/article.go` | Enrich with avatar/cover in PubList/PubDetail |
| `bff/handler/notification.go` | Add SourceAvatarUrl to NotificationVO |
| `bff/ioc/web.go` | Register upload handler |
| `bff/wire_gen.go` | Add upload handler wiring |
| `bff/config/docker.yaml` | Add OSS config |
| `docker-compose.infra.yaml` | Add RustFS service |
| `script/mysql/init.sql` | Add avatar_url/cover_url columns |

---

## Task 1: Update Proto Definitions

**Files:**
- Modify: `api/proto/user/v1/user.proto`
- Modify: `api/proto/article/v1/article.proto`
- Modify: `api/proto/notification/v2/notification.proto`

- [ ] **Step 1: Add avatar_url to User proto**

In `api/proto/user/v1/user.proto`, add to the `User` message (after `birthday` field 9):

```protobuf
  string avatar_url = 10;
```

- [ ] **Step 2: Add cover_url to Article proto**

In `api/proto/article/v1/article.proto`, add to the `Article` message (after `utime` field 7):

```protobuf
  string cover_url = 8;
```

- [ ] **Step 3: Add source_avatar_url to Notification proto**

In `api/proto/notification/v2/notification.proto`, add to the `NotificationItem` message (after `utime` field 20):

```protobuf
  string source_avatar_url = 21;
```

- [ ] **Step 4: Regenerate proto Go code**

Run: `cd /d/wen/demo/go/project/webook && make grpc`

- [ ] **Step 5: Verify generated code compiles**

Run: `go build ./api/proto/gen/...`

- [ ] **Step 6: Commit**

```bash
git add api/proto/
git commit -m "feat: add avatar_url, cover_url to proto definitions

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 2: Update User Microservice (domain + DAO + gRPC)

**Files:**
- Modify: `user/domain/user.go`
- Modify: `user/repository/dao/user.go`
- Modify: `user/grpc/server.go`

- [ ] **Step 1: Add AvatarUrl to user domain**

In `user/domain/user.go`, add `AvatarUrl` field to the `User` struct:

```go
type User struct {
	Id        int64
	Email     string
	Nickname  string
	Phone     string
	Password  string
	WechatInfo WechatInfo
	AboutMe   string
	AvatarUrl string
	Ctime     time.Time
	Birthday  time.Time
}
```

- [ ] **Step 2: Add AvatarUrl to user DAO struct**

In `user/repository/dao/user.go`, add `AvatarUrl` to the `User` struct:

```go
	AvatarUrl string `gorm:"type:varchar(512);default:''"`
```

Add it after the `AboutMe` field.

- [ ] **Step 3: Update user gRPC toDTO and toV conversions**

In `user/grpc/server.go`, update the `toDTO` method to include AvatarUrl:

In the `toDTO` function that converts proto User to domain User, add:
```go
	AvatarUrl: u.GetAvatarUrl(),
```

In the `toV` function that converts domain User to proto User, add:
```go
	AvatarUrl: u.AvatarUrl,
```

- [ ] **Step 4: Verify user service compiles**

Run: `go build ./user/...`

- [ ] **Step 5: Commit**

```bash
git add user/
git commit -m "feat(user): add avatar_url to domain, DAO, gRPC

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 3: Update Article Microservice (domain + DAO + gRPC)

**Files:**
- Modify: `article/domain/article.go`
- Modify: `article/repository/dao/gorm.go`
- Modify: `article/grpc/server.go`

- [ ] **Step 1: Add CoverUrl to article domain**

In `article/domain/article.go`, add `CoverUrl` to the `Article` struct:

```go
type Article struct {
	Id       int64
	Title    string
	Content  string
	CoverUrl string
	Author   Author
	Status   ArticleStatus
	Ctime    time.Time
	Utime    time.Time
}
```

- [ ] **Step 2: Add CoverUrl to article DAO structs**

In `article/repository/dao/gorm.go`, add `CoverUrl` to the `Article` struct:

```go
	CoverUrl string `gorm:"type:varchar(512);default:''" bson:"cover_url,omitempty"`
```

Add it after the `Content` field. Since `PublishedArticle` is defined as `type PublishedArticle Article`, it automatically gets the field.

- [ ] **Step 3: Update article gRPC toDTO and toV conversions**

In `article/grpc/server.go`:

In the `toDTO` function (proto → domain), add:
```go
	CoverUrl: art.GetCoverUrl(),
```

In the `toV` function (domain → proto), add:
```go
	CoverUrl: art.CoverUrl,
```

- [ ] **Step 4: Update article repository/service layer**

Check if article repository has conversion functions between DAO and domain. If so, add CoverUrl mapping there too. The DAO → domain and domain → DAO conversions need to carry CoverUrl through.

- [ ] **Step 5: Verify article service compiles**

Run: `go build ./article/...`

- [ ] **Step 6: Commit**

```bash
git add article/
git commit -m "feat(article): add cover_url to domain, DAO, gRPC

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 4: Add RustFS to Docker Infrastructure

**Files:**
- Modify: `docker-compose.infra.yaml`
- Modify: `script/mysql/init.sql`

- [ ] **Step 1: Add RustFS service to docker-compose.infra.yaml**

Add at the end of the `services` section (before any volumes section):

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
      - ./data/rustfs:/data
    command: server /data --console-address ":9001"
```

- [ ] **Step 2: Update init.sql with new columns**

Add to `script/mysql/init.sql`, after the users table creation:

```sql
-- Add avatar_url to users table (idempotent - won't fail if column exists)
ALTER TABLE `webook`.`users` ADD COLUMN IF NOT EXISTS `avatar_url` VARCHAR(512) DEFAULT '';

-- Add cover_url to articles tables
ALTER TABLE `webook`.`articles` ADD COLUMN IF NOT EXISTS `cover_url` VARCHAR(512) DEFAULT '';
ALTER TABLE `webook`.`published_articles` ADD COLUMN IF NOT EXISTS `cover_url` VARCHAR(512) DEFAULT '';
```

- [ ] **Step 3: Commit**

```bash
git add docker-compose.infra.yaml script/mysql/init.sql
git commit -m "infra: add RustFS service and image DB columns

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 5: Create BFF Upload Handler and S3 Client

**Files:**
- Create: `bff/ioc/oss.go`
- Create: `bff/handler/upload.go`
- Modify: `bff/config/docker.yaml`

- [ ] **Step 1: Add OSS config to BFF config**

In `bff/config/docker.yaml`, add an `oss` section:

```yaml
oss:
  endpoint: "http://localhost:9000"
  access_key: "admin"
  secret_key: "admin123456"
  region: "us-east-1"
  use_path_style: true
  avatar_bucket: "avatars"
  cover_bucket: "covers"
  image_bucket: "images"
```

- [ ] **Step 2: Create S3 client initialization**

Create `bff/ioc/oss.go`:

```go
package ioc

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/viper"
)

type OSSConfig struct {
	Endpoint     string `mapstructure:"endpoint"`
	AccessKey    string `mapstructure:"access_key"`
	SecretKey    string `mapstructure:"secret_key"`
	Region       string `mapstructure:"region"`
	UsePathStyle bool   `mapstructure:"use_path_style"`
	AvatarBucket string `mapstructure:"avatar_bucket"`
	CoverBucket  string `mapstructure:"cover_bucket"`
	ImageBucket  string `mapstructure:"image_bucket"`
}

func InitOSSClient() (*s3.S3, OSSConfig) {
	var cfg OSSConfig
	err := viper.UnmarshalKey("oss", &cfg)
	if err != nil {
		panic(err)
	}

	sess, err := session.NewSession(&aws.Config{
		Endpoint:         aws.String(cfg.Endpoint),
		Region:           aws.String(cfg.Region),
		Credentials:      credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey, ""),
		S3ForcePathStyle: aws.Bool(cfg.UsePathStyle),
	})
	if err != nil {
		panic(err)
	}

	return s3.New(sess), cfg
}
```

- [ ] **Step 3: Create upload handler**

Create `bff/handler/upload.go`:

```go
package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	userv1 "webook/api/proto/gen/user/v1"
	ijwt "webook/bff/handler/jwt"
	"webook/bff/ioc"
	"webook/pkg/ginx"
	"webook/pkg/logger"
)

type UploadHandler struct {
	s3Client *s3.S3
	ossCfg   ioc.OSSConfig
	userSvc  userv1.UserServiceClient
	l        logger.LoggerV1
}

func NewUploadHandler(s3Client *s3.S3, ossCfg ioc.OSSConfig, userSvc userv1.UserServiceClient, l logger.LoggerV1) *UploadHandler {
	return &UploadHandler{s3Client: s3Client, ossCfg: ossCfg, userSvc: userSvc, l: l}
}

func (h *UploadHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/upload")
	g.POST("/avatar", ginx.WrapClaims(h.l, h.UploadAvatar))
	g.POST("/image", ginx.WrapClaims(h.l, h.UploadImage))
	g.POST("/cover", ginx.WrapClaims(h.l, h.UploadCover))
}

var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

func (h *UploadHandler) UploadAvatar(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		return ginx.Result{Code: 4, Msg: "请选择文件"}, nil
	}
	defer file.Close()

	if header.Size > 2*1024*1024 {
		return ginx.Result{Code: 4, Msg: "头像文件不能超过2MB"}, nil
	}

	contentType := header.Header.Get("Content-Type")
	if !allowedImageTypes[contentType] {
		return ginx.Result{Code: 4, Msg: "只支持 jpg/png/webp/gif 格式"}, nil
	}

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	key := fmt.Sprintf("%d/avatar%s", uc.UserId, ext)

	data, err := io.ReadAll(file)
	if err != nil {
		return ginx.Result{Code: 5, Msg: "读取文件失败"}, err
	}

	_, err = h.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(h.ossCfg.AvatarBucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "上传失败"}, err
	}

	url := fmt.Sprintf("%s/%s/%s", h.ossCfg.Endpoint, h.ossCfg.AvatarBucket, key)

	// Update user avatar_url
	_, err = h.userSvc.EditNoSensitive(ctx.Request.Context(), &userv1.EditNoSensitiveRequest{
		User: &userv1.User{
			Id:        uc.UserId,
			AvatarUrl: url,
		},
	})
	if err != nil {
		h.l.Error("更新头像URL失败", logger.Error(err))
	}

	return ginx.Result{Code: 2, Msg: "OK", Data: map[string]string{"url": url}}, nil
}

func (h *UploadHandler) UploadImage(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		return ginx.Result{Code: 4, Msg: "请选择文件"}, nil
	}
	defer file.Close()

	if header.Size > 10*1024*1024 {
		return ginx.Result{Code: 4, Msg: "图片文件不能超过10MB"}, nil
	}

	contentType := header.Header.Get("Content-Type")
	if !allowedImageTypes[contentType] {
		return ginx.Result{Code: 4, Msg: "只支持 jpg/png/webp/gif 格式"}, nil
	}

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	key := fmt.Sprintf("articles/%s%s", uuid.New().String(), ext)

	data, err := io.ReadAll(file)
	if err != nil {
		return ginx.Result{Code: 5, Msg: "读取文件失败"}, err
	}

	_, err = h.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(h.ossCfg.ImageBucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "上传失败"}, err
	}

	url := fmt.Sprintf("%s/%s/%s", h.ossCfg.Endpoint, h.ossCfg.ImageBucket, key)
	return ginx.Result{Code: 2, Msg: "OK", Data: map[string]string{"url": url}}, nil
}

func (h *UploadHandler) UploadCover(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		return ginx.Result{Code: 4, Msg: "请选择文件"}, nil
	}
	defer file.Close()

	articleIdStr := ctx.Query("article_id")
	articleId, _ := strconv.ParseInt(articleIdStr, 10, 64)
	if articleId <= 0 {
		return ginx.Result{Code: 4, Msg: "article_id 参数错误"}, nil
	}

	if header.Size > 5*1024*1024 {
		return ginx.Result{Code: 4, Msg: "封面文件不能超过5MB"}, nil
	}

	contentType := header.Header.Get("Content-Type")
	if !allowedImageTypes[contentType] {
		return ginx.Result{Code: 4, Msg: "只支持 jpg/png/webp/gif 格式"}, nil
	}

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	key := fmt.Sprintf("articles/%d/cover%s", articleId, ext)

	data, err := io.ReadAll(file)
	if err != nil {
		return ginx.Result{Code: 5, Msg: "读取文件失败"}, err
	}

	_, err = h.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(h.ossCfg.CoverBucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "上传失败"}, err
	}

	url := fmt.Sprintf("%s/%s/%s", h.ossCfg.Endpoint, h.ossCfg.CoverBucket, key)
	return ginx.Result{Code: 2, Msg: "OK", Data: map[string]string{"url": url}}, nil
}
```

Note: `strings` import may not be needed — remove unused imports after writing.

- [ ] **Step 4: Add `github.com/google/uuid` dependency**

Run: `cd /d/wen/demo/go/project/webook && go get github.com/google/uuid`

- [ ] **Step 5: Commit**

```bash
git add bff/ioc/oss.go bff/handler/upload.go bff/config/docker.yaml go.mod go.sum
git commit -m "feat(bff): add image upload handler with RustFS S3

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 6: Update BFF VOs and Enrichment

**Files:**
- Modify: `bff/handler/article_vo.go`
- Modify: `bff/handler/user.go`
- Modify: `bff/handler/article.go`
- Modify: `bff/handler/notification.go`

- [ ] **Step 1: Add image fields to ArticleVO and Comment**

In `bff/handler/article_vo.go`:

Add to `ArticleVO` struct:
```go
	CoverUrl        string `json:"cover_url,omitempty"`
	AuthorAvatarUrl string `json:"author_avatar_url,omitempty"`
```

Add to `Comment` struct:
```go
	UserAvatarUrl string `json:"user_avatar_url,omitempty"`
```

- [ ] **Step 2: Add AvatarUrl to user profile VOs**

In `bff/handler/user.go`:

Add `AvatarUrl string` to the `Profile` struct (the one returned by self profile):
```go
	AvatarUrl string `json:"avatar_url"`
```

Add `AvatarUrl string` to `PublicProfile` struct:
```go
	AvatarUrl string `json:"avatar_url"`
```

Add `AvatarUrl string` to `RecommendUserVO` struct:
```go
	AvatarUrl string `json:"avatar_url"`
```

- [ ] **Step 3: Update Profile handler to include AvatarUrl**

In the `Profile` handler (self profile), where the response `Profile` struct is populated, add:
```go
	AvatarUrl: userResp.GetUser().GetAvatarUrl(),
```

In `ProfileById` handler (public profile), where `PublicProfile` is populated, add:
```go
	AvatarUrl: userResp.GetUser().GetAvatarUrl(),
```

In the `Recommend` handler, where `RecommendUserVO` is populated, add:
```go
	AvatarUrl: userResp.GetUser().GetAvatarUrl(),
```

- [ ] **Step 4: Update PubDetail to enrich with avatar and cover**

In `bff/handler/article.go`, in the `PubDetail` method:

Where `ArticleVO` is constructed, add:
```go
	CoverUrl:        art.GetCoverUrl(),    // from article proto
	AuthorAvatarUrl: userResp.GetUser().GetAvatarUrl(),  // from user profile fetch
```

The `userResp` is already fetched in PubDetail for `authorName`.

- [ ] **Step 5: Update PubList to enrich with avatar and cover**

In the `PubList` method, where `ArticleVO` is constructed inside the loop, add:
```go
	CoverUrl:        art.GetCoverUrl(),
	AuthorAvatarUrl: authorAvatarUrl,  // extract from userResp
```

Where `authorName` is extracted from `userResp`, also extract avatar:
```go
	var authorAvatarUrl string
	if userResp != nil && userResp.GetUser() != nil {
		authorName = userResp.GetUser().NickName
		authorAvatarUrl = userResp.GetUser().GetAvatarUrl()
	}
```

- [ ] **Step 6: Update GetComment to enrich with user avatar**

In the `GetComment` method, where `Comment` VOs are constructed:

Currently comments only have `UserName` populated by looking up user profiles. Add `UserAvatarUrl` from the same user profile fetch:
```go
	UserAvatarUrl: userResp.GetUser().GetAvatarUrl(),
```

Do the same for `GetMoreReplies`.

- [ ] **Step 7: Add SourceAvatarUrl to NotificationVO**

In `bff/handler/notification.go`:

Add to `NotificationVO` struct:
```go
	SourceAvatarUrl string `json:"source_avatar_url,omitempty"`
```

In the `toVOs` conversion function, add:
```go
	SourceAvatarUrl: n.GetSourceAvatarUrl(),
```

- [ ] **Step 8: Commit**

```bash
git add bff/handler/article_vo.go bff/handler/user.go bff/handler/article.go bff/handler/notification.go
git commit -m "feat(bff): add avatar/cover to all VOs and enrichment

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 7: Wire Up Upload Handler and Final Build

**Files:**
- Modify: `bff/ioc/web.go`
- Modify: `bff/wire_gen.go`

- [ ] **Step 1: Add upload handler to InitGin**

In `bff/ioc/web.go`, add `uploadHdl *handler.UploadHandler` parameter to `InitGin` and register routes:

```go
func InitGin(mdls []gin.HandlerFunc, hdl *handler.UserHandler, ..., uploadHdl *handler.UploadHandler) *gin.Engine {
	// ... existing handler registrations ...
	uploadHdl.RegisterRoutes(server)
	return server
}
```

- [ ] **Step 2: Update wire_gen.go**

In `bff/wire_gen.go`, add the upload handler initialization:

```go
	s3Client, ossCfg := ioc.InitOSSClient()
	uploadHandler := handler.NewUploadHandler(s3Client, ossCfg, userServiceClient, loggerV1)
```

And add `uploadHandler` to the `InitGin` call.

- [ ] **Step 3: Verify full BFF build**

Run: `cd /d/wen/demo/go/project/webook && go build ./bff/...`

- [ ] **Step 4: Verify user and article services build**

Run: `go build ./user/... && go build ./article/...`

- [ ] **Step 5: Commit**

```bash
git add bff/
git commit -m "feat(bff): wire upload handler into DI

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 8: Final Verification

- [ ] **Step 1: Full project build**

Run: `cd /d/wen/demo/go/project/webook && go build ./...`

Expected: BUILD SUCCESS (or only pre-existing unrelated errors)

- [ ] **Step 2: Go vet on changed packages**

Run: `go vet ./user/... ./article/... ./bff/...`

- [ ] **Step 3: Final commit**

```bash
git add -A
git commit -m "feat: image system complete - avatar, cover, upload

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```
