package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	articlev1 "webook/api/proto/gen/article/v1"
	tagv1 "webook/api/proto/gen/tag/v1"
	ijwt "webook/bff/handler/jwt"
	"webook/pkg/ginx"
	"webook/pkg/logger"
)

type TagHandler struct {
	svc        tagv1.TagServiceClient
	articleSvc articlev1.ArticleServiceClient
	l          logger.LoggerV1
}

func NewTagHandler(svc tagv1.TagServiceClient, articleSvc articlev1.ArticleServiceClient, l logger.LoggerV1) *TagHandler {
	return &TagHandler{svc: svc, articleSvc: articleSvc, l: l}
}

func (h *TagHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/tags")
	g.GET("", ginx.Wrap(h.l, h.GetTags))                        // 获取全局标签池
	g.POST("", ginx.WrapBodyAndToken(h.l, h.CreateTag))          // 创建标签
	g.POST("/attach", ginx.WrapBodyAndToken(h.l, h.AttachTags))  // 为资源绑定标签
	g.GET("/detail", ginx.WrapBody(h.l, h.GetTagById))           // 获取标签详情（话题详情页）
	g.GET("/biz", ginx.WrapBody(h.l, h.GetBizTags))              // 获取资源的标签
	g.GET("/biz_ids", ginx.WrapBody(h.l, h.GetBizIdsByTag))      // 按标签获取文章ID列表
	g.GET("/biz_count", ginx.WrapBody(h.l, h.CountBizByTag))     // 统计标签下的文章数量
	g.POST("/batch_biz", ginx.WrapBody(h.l, h.BatchGetBizTags))
	g.GET("/followed", ginx.WrapClaims(h.l, h.GetFollowedTags))
	g.POST("/:id/follow", ginx.WrapClaims(h.l, h.FollowTag))
	g.DELETE("/:id/follow", ginx.WrapClaims(h.l, h.UnfollowTag))
	g.GET("/:id/follow", ginx.WrapClaims(h.l, h.CheckTagFollow))
}

type CreateTagReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TagVO struct {
	Id            int64  `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	FollowerCount int64  `json:"follower_count,omitempty"`
}

// GetTags 获取全局标签池
func (h *TagHandler) GetTags(ctx *gin.Context) (ginx.Result, error) {
	resp, err := h.svc.GetTags(ctx.Request.Context(), &tagv1.GetTagsRequest{})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
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
	return ginx.Result{
		Code: 2,
		Msg:  "OK",
		Data: tags,
	}, nil
}

// CreateTag 创建标签
func (h *TagHandler) CreateTag(ctx *gin.Context, req CreateTagReq, uc ijwt.UserClaims) (ginx.Result, error) {
	if req.Name == "" {
		return ginx.Result{Code: 4, Msg: "标签名不能为空"}, nil
	}

	resp, err := h.svc.CreateTag(ctx.Request.Context(), &tagv1.CreateTagRequest{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "创建标签失败"}, err
	}

	return ginx.Result{
		Code: 2,
		Msg:  "OK",
		Data: TagVO{
			Id:          resp.GetTag().GetId(),
			Name:        resp.GetTag().GetName(),
			Description: resp.GetTag().GetDescription(),
		},
	}, nil
}

type AttachTagsReq struct {
	TagIds []int64 `json:"tag_ids"`
	Biz    string  `json:"biz"`
	BizId  int64   `json:"biz_id"`
}

// AttachTags 为资源绑定标签
func (h *TagHandler) AttachTags(ctx *gin.Context, req AttachTagsReq, uc ijwt.UserClaims) (ginx.Result, error) {
	if req.Biz == "" || req.BizId == 0 {
		return ginx.Result{Code: 4, Msg: "参数错误"}, nil
	}

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

type GetTagByIdReq struct {
	Id int64 `form:"id"`
}

// GetTagById 获取单个标签详情（话题详情页使用）
func (h *TagHandler) GetTagById(ctx *gin.Context, req GetTagByIdReq) (ginx.Result, error) {
	if req.Id == 0 {
		return ginx.Result{Code: 4, Msg: "参数错误"}, nil
	}

	resp, err := h.svc.GetTagById(ctx.Request.Context(), &tagv1.GetTagByIdRequest{
		Id: req.Id,
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "获取标签失败"}, err
	}

	return ginx.Result{
		Code: 2,
		Msg:  "OK",
		Data: TagVO{
			Id:            resp.GetTag().GetId(),
			Name:          resp.GetTag().GetName(),
			Description:   resp.GetTag().GetDescription(),
			FollowerCount: resp.GetTag().GetFollowerCount(),
		},
	}, nil
}

type GetBizTagsReq struct {
	Biz   string `form:"biz"`
	BizId int64  `form:"biz_id"`
}

// GetBizTags 获取资源的标签
func (h *TagHandler) GetBizTags(ctx *gin.Context, req GetBizTagsReq) (ginx.Result, error) {
	if req.Biz == "" || req.BizId == 0 {
		return ginx.Result{Code: 4, Msg: "参数错误"}, nil
	}

	resp, err := h.svc.GetBizTags(ctx.Request.Context(), &tagv1.GetBizTagsRequest{
		Biz:   req.Biz,
		BizId: req.BizId,
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "获取标签失败"}, err
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
	return ginx.Result{
		Code: 2,
		Msg:  "OK",
		Data: tags,
	}, nil
}

type GetBizIdsByTagReq struct {
	Biz    string `form:"biz"`
	TagId  int64  `form:"tag_id"`
	Offset int32  `form:"offset"`
	Limit  int32  `form:"limit"`
	SortBy string `form:"sort_by"` // newest, hottest, featured
}

// GetBizIdsByTag 按标签获取文章ID列表（支持排序）
func (h *TagHandler) GetBizIdsByTag(ctx *gin.Context, req GetBizIdsByTagReq) (ginx.Result, error) {
	if req.Biz == "" || req.TagId == 0 {
		return ginx.Result{Code: 4, Msg: "参数错误"}, nil
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	if req.SortBy == "" {
		req.SortBy = "newest"
	}

	resp, err := h.svc.GetBizIdsByTag(ctx.Request.Context(), &tagv1.GetBizIdsByTagRequest{
		Biz:    req.Biz,
		TagId:  req.TagId,
		Offset: req.Offset,
		Limit:  req.Limit,
		SortBy: req.SortBy,
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "获取文章列表失败"}, err
	}

	return ginx.Result{
		Code: 2,
		Msg:  "OK",
		Data: resp.GetBizIds(),
	}, nil
}

type CountBizByTagReq struct {
	Biz   string `form:"biz"`
	TagId int64  `form:"tag_id"`
}

// CountBizByTag 统计标签下的文章数量
func (h *TagHandler) CountBizByTag(ctx *gin.Context, req CountBizByTagReq) (ginx.Result, error) {
	if req.Biz == "" || req.TagId == 0 {
		return ginx.Result{Code: 4, Msg: "参数错误"}, nil
	}

	resp, err := h.svc.CountBizByTag(ctx.Request.Context(), &tagv1.CountBizByTagRequest{
		Biz:   req.Biz,
		TagId: req.TagId,
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "获取统计失败"}, err
	}

	return ginx.Result{
		Code: 2,
		Msg:  "OK",
		Data: resp.GetCount(),
	}, nil
}

// FollowTag 关注标签
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

// UnfollowTag 取消关注标签
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

// CheckTagFollow 检查是否关注了标签
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

type GetFollowedTagsReq struct {
	Offset int32 `form:"offset"`
	Limit  int32 `form:"limit"`
}

// GetFollowedTags 获取用户关注的标签列表
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

type BatchGetBizTagsReq struct {
	Biz    string  `json:"biz"`
	BizIds []int64 `json:"biz_ids"`
}

// BatchGetBizTags 批量获取资源的标签
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
