# Tag Service Enhancements Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add tag follow system, batch tag queries, real hot/featured sorting, and attach permission checks to the existing tag service.

**Architecture:** Extends existing tag microservice (domain → dao → repo → service → grpc) bottom-up, then adds BFF endpoints. Hot/featured sorting is computed in the service layer using interactive service data. Tag follow is a new DAO table with counter on tags table.

**Tech Stack:** Go, gRPC/protobuf, GORM, Redis, Gin, Kafka

**Spec:** `docs/superpowers/specs/2026-03-21-tag-service-enhancements-design.md`

---

## File Structure

### New Files
_None — all changes are modifications to existing files._

### Modified Files

| File | Responsibility |
|------|---------------|
| `api/proto/tag/v1/tag.proto` | Add follower_count to Tag, add 5 new RPCs |
| `tag/domain/tag.go` | Add FollowerCount field |
| `tag/repository/dao/types.go` | Add TagFollow struct, FollowerCount to Tag, BizCtime to TagBiz, 5 new DAO methods |
| `tag/repository/dao/gorm.go` | Implement 5 new DAO methods |
| `tag/repository/types.go` | Add 5 new repository interface methods |
| `tag/repository/tag.go` | Implement 5 new repository methods, update toDomain/toEntity |
| `tag/service/types.go` | Add 5 new service interface methods |
| `tag/service/tag.go` | Implement follow logic + hot/featured sorting with intrSvc |
| `tag/grpc/tag.go` | Implement 5 new gRPC handlers, update toDTO |
| `bff/handler/tag.go` | Add 5 new HTTP routes, permission check on AttachTags |

### Proto Regeneration (after Task 1)
| File | Action |
|------|--------|
| `api/proto/gen/tag/v1/tag.pb.go` | Regenerate with `make grpc` |
| `api/proto/gen/tag/v1/tag_grpc.pb.go` | Regenerate with `make grpc` |

---

## Task 1: Update Proto Definition

**Files:**
- Modify: `api/proto/tag/v1/tag.proto`

- [ ] **Step 1: Add follower_count to Tag message and add 5 new RPCs**

In `api/proto/tag/v1/tag.proto`, add to the `service TagService` block (after `CountBizByTag`):

```protobuf
  // 关注标签
  rpc FollowTag(FollowTagRequest) returns (FollowTagResponse);
  // 取消关注标签
  rpc UnfollowTag(UnfollowTagRequest) returns (UnfollowTagResponse);
  // 检查是否关注了标签
  rpc CheckTagFollow(CheckTagFollowRequest) returns (CheckTagFollowResponse);
  // 获取用户关注的标签列表
  rpc GetUserFollowedTags(GetUserFollowedTagsRequest) returns (GetUserFollowedTagsResponse);
  // 批量获取资源的标签
  rpc BatchGetBizTags(BatchGetBizTagsRequest) returns (BatchGetBizTagsResponse);
```

Add `follower_count` to the existing `Tag` message:

```protobuf
message Tag {
  int64 id = 1;
  string name = 2;
  string description = 3;
  int64 follower_count = 4;
}
```

Add new messages at the end of the file:

```protobuf
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

- [ ] **Step 2: Regenerate proto Go code**

Run: `cd D:\wen\demo\go\project\webook && make grpc`

If `make grpc` is not available or fails, run directly:

```bash
protoc --go_out=. --go-grpc_out=. api/proto/tag/v1/tag.proto
```

Expected: `api/proto/gen/tag/v1/tag.pb.go` and `tag_grpc.pb.go` regenerated with new types and interfaces.

- [ ] **Step 3: Verify regenerated code compiles**

Run: `cd D:\wen\demo\go\project\webook && go build ./api/proto/gen/tag/v1/...`

Expected: BUILD SUCCESS (the tag service itself won't compile yet because the new interface methods are unimplemented).

- [ ] **Step 4: Commit**

```bash
git add api/proto/tag/v1/tag.proto api/proto/gen/tag/v1/
git commit -m "feat(tag): add follow, batch, sort proto definitions"
```

---

## Task 2: Update Domain and DAO Layer

**Files:**
- Modify: `tag/domain/tag.go`
- Modify: `tag/repository/dao/types.go`
- Modify: `tag/repository/dao/gorm.go`

- [ ] **Step 1: Add FollowerCount to domain**

In `tag/domain/tag.go`, add `FollowerCount` field:

```go
type Tag struct {
	Id            int64
	Name          string
	Description   string
	FollowerCount int64
}
```

- [ ] **Step 2: Update DAO types — add TagFollow, FollowerCount, BizCtime, new methods**

In `tag/repository/dao/types.go`:

Add `FollowerCount` to `Tag` struct:
```go
type Tag struct {
	Id            int64  `gorm:"primaryKey,autoIncrement"`
	Name          string `gorm:"type:varchar(256);uniqueIndex"`
	Description   string `gorm:"type:varchar(1024)"`
	FollowerCount int64  `gorm:"default:0"`
	Ctime         int64
	Utime         int64
}
```

Add `BizCtime` to `TagBiz` struct:
```go
type TagBiz struct {
	Id       int64  `gorm:"primaryKey,autoIncrement"`
	BizId    int64  `gorm:"uniqueIndex:biz_type_id_tid"`
	Biz      string `gorm:"type:varchar(128);uniqueIndex:biz_type_id_tid"`
	Tid      int64  `gorm:"uniqueIndex:biz_type_id_tid"`
	Tag      *Tag   `gorm:"foreignKey:Tid;AssociationForeignKey:Id;constraint:OnDelete:CASCADE"`
	BizCtime int64  `gorm:"default:0"`
	Ctime    int64
	Utime    int64
}
```

Add new `TagFollow` struct:
```go
type TagFollow struct {
	Id    int64 `gorm:"primaryKey,autoIncrement"`
	Uid   int64 `gorm:"uniqueIndex:uk_uid_tag"`
	TagId int64 `gorm:"uniqueIndex:uk_uid_tag;index:idx_tag_id"`
	Ctime int64
}
```

Add 5 new methods to `TagDao` interface:
```go
type TagDao interface {
	// ... existing methods ...
	FollowTag(ctx context.Context, uid, tagId int64) error
	UnfollowTag(ctx context.Context, uid, tagId int64) error
	CheckTagFollow(ctx context.Context, uid, tagId int64) (bool, error)
	GetUserFollowedTags(ctx context.Context, uid int64, offset, limit int) ([]Tag, error)
	BatchGetTagsByBiz(ctx context.Context, biz string, bizIds []int64) (map[int64][]Tag, error)
}
```

- [ ] **Step 3: Implement 5 new DAO methods in gorm.go**

Add to `tag/repository/dao/gorm.go`:

```go
func (d *GORMTagDao) FollowTag(ctx context.Context, uid, tagId int64) error {
	now := time.Now().UnixMilli()
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Where("uid = ? AND tag_id = ?", uid, tagId).FirstOrCreate(&TagFollow{
			Uid:   uid,
			TagId: tagId,
			Ctime: now,
		})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			// already followed, idempotent
			return nil
		}
		return tx.Model(&Tag{}).Where("id = ?", tagId).
			UpdateColumn("follower_count", gorm.Expr("follower_count + 1")).Error
	})
}

func (d *GORMTagDao) UnfollowTag(ctx context.Context, uid, tagId int64) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Where("uid = ? AND tag_id = ?", uid, tagId).Delete(&TagFollow{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			// not followed, idempotent
			return nil
		}
		return tx.Model(&Tag{}).Where("id = ? AND follower_count > 0", tagId).
			UpdateColumn("follower_count", gorm.Expr("follower_count - 1")).Error
	})
}

func (d *GORMTagDao) CheckTagFollow(ctx context.Context, uid, tagId int64) (bool, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&TagFollow{}).
		Where("uid = ? AND tag_id = ?", uid, tagId).
		Count(&count).Error
	return count > 0, err
}

func (d *GORMTagDao) GetUserFollowedTags(ctx context.Context, uid int64, offset, limit int) ([]Tag, error) {
	var tags []Tag
	err := d.db.WithContext(ctx).
		Joins("JOIN tag_follows tf ON tags.id = tf.tag_id").
		Where("tf.uid = ?", uid).
		Order("tf.ctime DESC").
		Offset(offset).Limit(limit).
		Find(&tags).Error
	return tags, err
}

func (d *GORMTagDao) BatchGetTagsByBiz(ctx context.Context, biz string, bizIds []int64) (map[int64][]Tag, error) {
	var tagBizs []TagBiz
	err := d.db.WithContext(ctx).Model(&TagBiz{}).
		InnerJoins("Tag").
		Where("biz = ? AND biz_id IN ?", biz, bizIds).
		Find(&tagBizs).Error
	if err != nil {
		return nil, err
	}
	result := make(map[int64][]Tag, len(bizIds))
	for _, tb := range tagBizs {
		if tb.Tag != nil {
			result[tb.BizId] = append(result[tb.BizId], *tb.Tag)
		}
	}
	return result, nil
}
```

Also ensure GORM auto-migrates the new `TagFollow` table. Check if there's an `InitTable` function in the DAO or if AutoMigrate is called elsewhere. If needed, add `&TagFollow{}` to the AutoMigrate list in the tag service's initialization.

- [ ] **Step 4: Commit**

```bash
git add tag/domain/tag.go tag/repository/dao/types.go tag/repository/dao/gorm.go
git commit -m "feat(tag): add follow DAO, batch query, domain updates"
```

---

## Task 3: Update Repository Layer

**Files:**
- Modify: `tag/repository/types.go`
- Modify: `tag/repository/tag.go`

- [ ] **Step 1: Add 5 new methods to repository interface**

In `tag/repository/types.go`, add to `TagRepository`:

```go
type TagRepository interface {
	// ... existing methods ...
	FollowTag(ctx context.Context, uid, tagId int64) error
	UnfollowTag(ctx context.Context, uid, tagId int64) error
	CheckTagFollow(ctx context.Context, uid, tagId int64) (bool, error)
	GetUserFollowedTags(ctx context.Context, uid int64, offset, limit int) ([]domain.Tag, error)
	BatchGetBizTags(ctx context.Context, biz string, bizIds []int64) (map[int64][]domain.Tag, error)
}
```

- [ ] **Step 2: Implement 5 new methods in CachedTagRepository**

In `tag/repository/tag.go`, add:

```go
func (r *CachedTagRepository) FollowTag(ctx context.Context, uid, tagId int64) error {
	return r.dao.FollowTag(ctx, uid, tagId)
}

func (r *CachedTagRepository) UnfollowTag(ctx context.Context, uid, tagId int64) error {
	return r.dao.UnfollowTag(ctx, uid, tagId)
}

func (r *CachedTagRepository) CheckTagFollow(ctx context.Context, uid, tagId int64) (bool, error) {
	return r.dao.CheckTagFollow(ctx, uid, tagId)
}

func (r *CachedTagRepository) GetUserFollowedTags(ctx context.Context, uid int64, offset, limit int) ([]domain.Tag, error) {
	tags, err := r.dao.GetUserFollowedTags(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Tag, 0, len(tags))
	for _, tag := range tags {
		res = append(res, r.toDomain(tag))
	}
	return res, nil
}

func (r *CachedTagRepository) BatchGetBizTags(ctx context.Context, biz string, bizIds []int64) (map[int64][]domain.Tag, error) {
	tagMap, err := r.dao.BatchGetTagsByBiz(ctx, biz, bizIds)
	if err != nil {
		return nil, err
	}
	result := make(map[int64][]domain.Tag, len(tagMap))
	for bizId, tags := range tagMap {
		domainTags := make([]domain.Tag, 0, len(tags))
		for _, tag := range tags {
			domainTags = append(domainTags, r.toDomain(tag))
		}
		result[bizId] = domainTags
	}
	return result, nil
}
```

Also update `toDomain` to include `FollowerCount`:

```go
func (r *CachedTagRepository) toDomain(tag dao.Tag) domain.Tag {
	return domain.Tag{
		Id:            tag.Id,
		Name:          tag.Name,
		Description:   tag.Description,
		FollowerCount: tag.FollowerCount,
	}
}
```

- [ ] **Step 3: Commit**

```bash
git add tag/repository/types.go tag/repository/tag.go
git commit -m "feat(tag): add follow and batch repository methods"
```

---

## Task 4: Update Service Layer (Follow + Sorting)

**Files:**
- Modify: `tag/service/types.go`
- Modify: `tag/service/tag.go`

- [ ] **Step 1: Add 5 new methods to service interface**

In `tag/service/types.go`:

```go
type TagService interface {
	// ... existing methods ...
	FollowTag(ctx context.Context, uid, tagId int64) error
	UnfollowTag(ctx context.Context, uid, tagId int64) error
	CheckTagFollow(ctx context.Context, uid, tagId int64) (bool, error)
	GetUserFollowedTags(ctx context.Context, uid int64, offset, limit int) ([]domain.Tag, error)
	BatchGetBizTags(ctx context.Context, biz string, bizIds []int64) (map[int64][]domain.Tag, error)
}
```

- [ ] **Step 2: Update tagService struct to inject intrSvc**

In `tag/service/tag.go`, update the struct and constructor:

```go
import (
	intrv1 "webook/api/proto/gen/intr/v1"
	// ... other imports ...
)

type tagService struct {
	repo     repository.TagRepository
	producer events.Producer
	intrSvc  intrv1.InteractiveServiceClient
	l        logger.LoggerV1
}

func NewTagService(repo repository.TagRepository, producer events.Producer, intrSvc intrv1.InteractiveServiceClient, l logger.LoggerV1) TagService {
	return &tagService{repo: repo, producer: producer, intrSvc: intrSvc, l: l}
}
```

- [ ] **Step 3: Implement follow methods**

```go
func (s *tagService) FollowTag(ctx context.Context, uid, tagId int64) error {
	return s.repo.FollowTag(ctx, uid, tagId)
}

func (s *tagService) UnfollowTag(ctx context.Context, uid, tagId int64) error {
	return s.repo.UnfollowTag(ctx, uid, tagId)
}

func (s *tagService) CheckTagFollow(ctx context.Context, uid, tagId int64) (bool, error) {
	return s.repo.CheckTagFollow(ctx, uid, tagId)
}

func (s *tagService) GetUserFollowedTags(ctx context.Context, uid int64, offset, limit int) ([]domain.Tag, error) {
	return s.repo.GetUserFollowedTags(ctx, uid, offset, limit)
}

func (s *tagService) BatchGetBizTags(ctx context.Context, biz string, bizIds []int64) (map[int64][]domain.Tag, error) {
	return s.repo.BatchGetBizTags(ctx, biz, bizIds)
}
```

- [ ] **Step 4: Update GetBizIdsByTag with real hot/featured sorting**

Replace the existing `GetBizIdsByTag` method:

```go
func (s *tagService) GetBizIdsByTag(ctx context.Context, biz string, tagId int64, offset, limit int, sortBy string) ([]int64, error) {
	if sortBy == "newest" || sortBy == "" {
		return s.repo.GetBizIdsByTag(ctx, biz, tagId, offset, limit, "newest")
	}

	// For hottest/featured, fetch all biz_ids (capped at 1000) then sort in memory
	allIds, err := s.repo.GetBizIdsByTag(ctx, biz, tagId, 0, 1000, "newest")
	if err != nil {
		return nil, err
	}
	if len(allIds) == 0 {
		return []int64{}, nil
	}

	// Fetch interactive data
	intrResp, err := s.intrSvc.GetByIds(ctx, &intrv1.GetByIdsRequest{
		Biz:    biz,
		BizIds: allIds,
	})
	if err != nil {
		s.l.Error("获取互动数据失败，降级为newest排序", logger.Error(err))
		return s.repo.GetBizIdsByTag(ctx, biz, tagId, offset, limit, "newest")
	}

	type scored struct {
		bizId int64
		score float64
	}
	items := make([]scored, 0, len(allIds))

	now := time.Now().Unix()
	for _, id := range allIds {
		intr := intrResp.GetIntrs()[id]
		var sc float64
		if intr != nil {
			switch sortBy {
			case "hottest":
				sc = float64(intr.LikeCnt)*3 + float64(intr.CollectCnt)*5 + float64(intr.ReadCnt)*0.1
			case "featured":
				hours := float64(now-intr.GetReadCnt()) / 3600 // approximate with ctime stored elsewhere
				if hours < 0 {
					hours = 1
				}
				quality := float64(intr.LikeCnt)*3 + float64(intr.CollectCnt)*5
				sc = quality / math.Pow(hours+2, 1.5)
			}
		}
		items = append(items, scored{bizId: id, score: sc})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].score > items[j].score
	})

	// Apply pagination
	start := offset
	if start > len(items) {
		return []int64{}, nil
	}
	end := start + limit
	if end > len(items) {
		end = len(items)
	}
	result := make([]int64, 0, end-start)
	for _, item := range items[start:end] {
		result = append(result, item.bizId)
	}
	return result, nil
}
```

Add `math` and `sort` to imports.

- [ ] **Step 5: Commit**

```bash
git add tag/service/types.go tag/service/tag.go
git commit -m "feat(tag): add follow service, hot/featured sorting with intrSvc"
```

---

## Task 5: Update gRPC Server

**Files:**
- Modify: `tag/grpc/tag.go`

- [ ] **Step 1: Implement 5 new gRPC handler methods**

Add to `tag/grpc/tag.go`:

```go
func (t *TagServiceServer) FollowTag(ctx context.Context, request *tagv1.FollowTagRequest) (*tagv1.FollowTagResponse, error) {
	err := t.svc.FollowTag(ctx, request.Uid, request.TagId)
	return &tagv1.FollowTagResponse{}, err
}

func (t *TagServiceServer) UnfollowTag(ctx context.Context, request *tagv1.UnfollowTagRequest) (*tagv1.UnfollowTagResponse, error) {
	err := t.svc.UnfollowTag(ctx, request.Uid, request.TagId)
	return &tagv1.UnfollowTagResponse{}, err
}

func (t *TagServiceServer) CheckTagFollow(ctx context.Context, request *tagv1.CheckTagFollowRequest) (*tagv1.CheckTagFollowResponse, error) {
	followed, err := t.svc.CheckTagFollow(ctx, request.Uid, request.TagId)
	if err != nil {
		return nil, err
	}
	return &tagv1.CheckTagFollowResponse{Followed: followed}, nil
}

func (t *TagServiceServer) GetUserFollowedTags(ctx context.Context, request *tagv1.GetUserFollowedTagsRequest) (*tagv1.GetUserFollowedTagsResponse, error) {
	tags, err := t.svc.GetUserFollowedTags(ctx, request.Uid, int(request.Offset), int(request.Limit))
	if err != nil {
		return nil, err
	}
	res := make([]*tagv1.Tag, 0, len(tags))
	for _, tag := range tags {
		res = append(res, t.toDTO(tag))
	}
	return &tagv1.GetUserFollowedTagsResponse{Tags: res}, nil
}

func (t *TagServiceServer) BatchGetBizTags(ctx context.Context, request *tagv1.BatchGetBizTagsRequest) (*tagv1.BatchGetBizTagsResponse, error) {
	tagMap, err := t.svc.BatchGetBizTags(ctx, request.Biz, request.BizIds)
	if err != nil {
		return nil, err
	}
	result := make(map[int64]*tagv1.BizTagList, len(tagMap))
	for bizId, tags := range tagMap {
		list := make([]*tagv1.Tag, 0, len(tags))
		for _, tag := range tags {
			list = append(list, t.toDTO(tag))
		}
		result[bizId] = &tagv1.BizTagList{Tags: list}
	}
	return &tagv1.BatchGetBizTagsResponse{BizTags: result}, nil
}
```

- [ ] **Step 2: Update toDTO to include FollowerCount**

```go
func (t *TagServiceServer) toDTO(tag domain.Tag) *tagv1.Tag {
	return &tagv1.Tag{
		Id:            tag.Id,
		Name:          tag.Name,
		Description:   tag.Description,
		FollowerCount: tag.FollowerCount,
	}
}
```

- [ ] **Step 3: Verify tag service compiles**

Run: `cd D:\wen\demo\go\project\webook && go build ./tag/...`

Expected: BUILD SUCCESS

- [ ] **Step 4: Commit**

```bash
git add tag/grpc/tag.go
git commit -m "feat(tag): implement gRPC handlers for follow and batch"
```

---

## Task 6: Update BFF Handler

**Files:**
- Modify: `bff/handler/tag.go`

- [ ] **Step 1: Add articleSvc dependency and update constructor**

Update the struct and constructor:

```go
import (
	articlev1 "webook/api/proto/gen/article/v1"
	// ... existing imports ...
)

type TagHandler struct {
	svc        tagv1.TagServiceClient
	articleSvc articlev1.ArticleServiceClient
	l          logger.LoggerV1
}

func NewTagHandler(svc tagv1.TagServiceClient, articleSvc articlev1.ArticleServiceClient, l logger.LoggerV1) *TagHandler {
	return &TagHandler{svc: svc, articleSvc: articleSvc, l: l}
}
```

- [ ] **Step 2: Update TagVO to include FollowerCount**

```go
type TagVO struct {
	Id            int64  `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	FollowerCount int64  `json:"follower_count,omitempty"`
}
```

- [ ] **Step 3: Register 5 new routes**

In `RegisterRoutes`, add after existing routes:

```go
func (h *TagHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/tags")
	// ... existing routes ...
	g.POST("/batch_biz", ginx.WrapBody(h.l, h.BatchGetBizTags))
	g.GET("/followed", ginx.WrapClaims(h.l, h.GetFollowedTags))
	g.POST("/:id/follow", ginx.WrapClaims(h.l, h.FollowTag))
	g.DELETE("/:id/follow", ginx.WrapClaims(h.l, h.UnfollowTag))
	g.GET("/:id/follow", ginx.WrapClaims(h.l, h.CheckTagFollow))
}
```

**Important:** The `/followed` route MUST be registered BEFORE `/:id/follow` routes, otherwise `/tags/followed` would match `/:id` with id="followed".

- [ ] **Step 4: Implement FollowTag handler**

```go
func (h *TagHandler) FollowTag(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	idStr := ctx.Param("id")
	tagId, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || tagId <= 0 {
		return ginx.Result{Code: 4, Msg: "参数错误"}, nil
	}
	_, err = h.svc.FollowTag(ctx.Request.Context(), &tagv1.FollowTagRequest{
		Uid:   uc.UserId,
		TagId: tagId,
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "关注失败"}, err
	}
	return ginx.Result{Code: 2, Msg: "OK"}, nil
}
```

Add `"strconv"` to imports.

- [ ] **Step 5: Implement UnfollowTag handler**

```go
func (h *TagHandler) UnfollowTag(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	idStr := ctx.Param("id")
	tagId, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || tagId <= 0 {
		return ginx.Result{Code: 4, Msg: "参数错误"}, nil
	}
	_, err = h.svc.UnfollowTag(ctx.Request.Context(), &tagv1.UnfollowTagRequest{
		Uid:   uc.UserId,
		TagId: tagId,
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "取消关注失败"}, err
	}
	return ginx.Result{Code: 2, Msg: "OK"}, nil
}
```

- [ ] **Step 6: Implement CheckTagFollow handler**

```go
func (h *TagHandler) CheckTagFollow(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	idStr := ctx.Param("id")
	tagId, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || tagId <= 0 {
		return ginx.Result{Code: 4, Msg: "参数错误"}, nil
	}
	resp, err := h.svc.CheckTagFollow(ctx.Request.Context(), &tagv1.CheckTagFollowRequest{
		Uid:   uc.UserId,
		TagId: tagId,
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "查询失败"}, err
	}
	return ginx.Result{Code: 2, Msg: "OK", Data: resp.GetFollowed()}, nil
}
```

- [ ] **Step 7: Implement GetFollowedTags handler**

```go
type GetFollowedTagsReq struct {
	Offset int32 `form:"offset"`
	Limit  int32 `form:"limit"`
}

func (h *TagHandler) GetFollowedTags(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	var req GetFollowedTagsReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		return ginx.Result{Code: 4, Msg: "参数错误"}, nil
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	resp, err := h.svc.GetUserFollowedTags(ctx.Request.Context(), &tagv1.GetUserFollowedTagsRequest{
		Uid:    uc.UserId,
		Offset: req.Offset,
		Limit:  req.Limit,
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "获取关注标签失败"}, err
	}
	tags := make([]TagVO, 0, len(resp.GetTags()))
	for _, t := range resp.GetTags() {
		tags = append(tags, TagVO{
			Id:            t.Id,
			Name:          t.Name,
			Description:   t.Description,
			FollowerCount: t.FollowerCount,
		})
	}
	return ginx.Result{Code: 2, Msg: "OK", Data: tags}, nil
}
```

- [ ] **Step 8: Implement BatchGetBizTags handler**

```go
type BatchGetBizTagsReq struct {
	Biz    string  `json:"biz"`
	BizIds []int64 `json:"biz_ids"`
}

func (h *TagHandler) BatchGetBizTags(ctx *gin.Context, req BatchGetBizTagsReq) (ginx.Result, error) {
	if req.Biz == "" || len(req.BizIds) == 0 {
		return ginx.Result{Code: 4, Msg: "参数错误"}, nil
	}
	if len(req.BizIds) > 100 {
		return ginx.Result{Code: 4, Msg: "批量查询最多100条"}, nil
	}
	resp, err := h.svc.BatchGetBizTags(ctx.Request.Context(), &tagv1.BatchGetBizTagsRequest{
		Biz:    req.Biz,
		BizIds: req.BizIds,
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "批量获取标签失败"}, err
	}
	result := make(map[int64][]TagVO, len(resp.GetBizTags()))
	for bizId, list := range resp.GetBizTags() {
		tags := make([]TagVO, 0, len(list.GetTags()))
		for _, t := range list.GetTags() {
			tags = append(tags, TagVO{
				Id:   t.Id,
				Name: t.Name,
			})
		}
		result[bizId] = tags
	}
	return ginx.Result{Code: 2, Msg: "OK", Data: result}, nil
}
```

- [ ] **Step 9: Add permission check to AttachTags**

Replace the existing `AttachTags` method:

```go
func (h *TagHandler) AttachTags(ctx *gin.Context, req AttachTagsReq, uc ijwt.UserClaims) (ginx.Result, error) {
	if req.Biz == "" || req.BizId == 0 {
		return ginx.Result{Code: 4, Msg: "参数错误"}, nil
	}

	// Permission check: only article author can set tags
	if req.Biz == "article" {
		artResp, err := h.articleSvc.GetById(ctx.Request.Context(), &articlev1.GetByIdRequest{Id: req.BizId})
		if err != nil {
			return ginx.Result{Code: 5, Msg: "系统错误"}, err
		}
		if artResp.GetArticle().GetAuthor().GetId() != uc.UserId {
			return ginx.Result{Code: 4, Msg: "只有作者可以设置标签"}, nil
		}
	}

	_, err := h.svc.AttachTags(ctx.Request.Context(), &tagv1.AttachTagsRequest{
		Tids:  req.TagIds,
		Biz:   req.Biz,
		BizId: req.BizId,
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "绑定标签失败"}, err
	}

	return ginx.Result{Code: 2, Msg: "OK"}, nil
}
```

- [ ] **Step 10: Update all TagVO usages in existing handlers to include FollowerCount**

In `GetTagById` handler, update the response:

```go
Data: TagVO{
	Id:            resp.GetTag().GetId(),
	Name:          resp.GetTag().GetName(),
	Description:   resp.GetTag().GetDescription(),
	FollowerCount: resp.GetTag().GetFollowerCount(),
},
```

In `GetTags` and `GetBizTags` loop, add FollowerCount:

```go
tags = append(tags, TagVO{
	Id:            t.Id,
	Name:          t.Name,
	Description:   t.Description,
	FollowerCount: t.FollowerCount,
})
```

- [ ] **Step 11: Commit**

```bash
git add bff/handler/tag.go
git commit -m "feat(bff/tag): add follow routes, batch query, permission check"
```

---

## Task 7: Update Dependency Injection

**Files:**
- Modify: Wherever `NewTagService` and `NewTagHandler` are called (wire files, ioc files)

- [ ] **Step 1: Update tag service NewTagService call to pass intrSvc**

Find where `NewTagService` is called (likely in `tag/wire_gen.go` or an ioc file) and add the `intrSvc` parameter. The tag service needs an `intrv1.InteractiveServiceClient` injected.

- [ ] **Step 2: Update BFF NewTagHandler call to pass articleSvc**

Find where `NewTagHandler` is called in BFF (likely `bff/ioc/web.go` or `bff/wire_gen.go`) and add the `articleSvc` parameter.

- [ ] **Step 3: Verify full build**

Run: `cd D:\wen\demo\go\project\webook && go build ./tag/... && go build ./bff/...`

Expected: Both BUILD SUCCESS

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "feat(tag): update dependency injection for follow and sorting"
```

---

## Task 8: Final Verification

- [ ] **Step 1: Verify full project builds**

Run: `cd D:\wen\demo\go\project\webook && go build ./...`

Expected: BUILD SUCCESS (or only unrelated errors from other services)

- [ ] **Step 2: Verify go vet passes**

Run: `cd D:\wen\demo\go\project\webook && go vet ./tag/... ./bff/...`

Expected: No issues

- [ ] **Step 3: Final commit**

```bash
git add -A
git commit -m "feat(tag): tag service enhancements complete - follow, batch, sorting, permissions"
```
