package handler

import (
	"encoding/json"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	articlev1 "webook/api/proto/gen/article/v1"
	commentv1 "webook/api/proto/gen/comment/v1"
	feedv1 "webook/api/proto/gen/feed"
	intrv1 "webook/api/proto/gen/intr/v1"
	tagv1 "webook/api/proto/gen/tag/v1"
	userv1 "webook/api/proto/gen/user/v1"
	ijwt "webook/bff/handler/jwt"
	"webook/pkg/ginx"
	"webook/pkg/logger"
)

type FeedHandler struct {
	svc        feedv1.FeedSvcClient
	articleSvc articlev1.ArticleServiceClient
	userSvc    userv1.UserServiceClient
	intrSvc    intrv1.InteractiveServiceClient
	commentSvc commentv1.CommentServiceClient
	tagSvc     tagv1.TagServiceClient
	l          logger.LoggerV1
}

func NewFeedHandler(svc feedv1.FeedSvcClient, articleSvc articlev1.ArticleServiceClient, userSvc userv1.UserServiceClient, intrSvc intrv1.InteractiveServiceClient, commentSvc commentv1.CommentServiceClient, tagSvc tagv1.TagServiceClient, l logger.LoggerV1) *FeedHandler {
	return &FeedHandler{
		svc:        svc,
		articleSvc: articleSvc,
		userSvc:    userSvc,
		intrSvc:    intrSvc,
		commentSvc: commentSvc,
		tagSvc:     tagSvc,
		l:          l,
	}
}

func (h *FeedHandler) RegisterRoutes(server *gin.Engine) {
	server.GET("/feed", ginx.WrapBodyAndToken(h.l, h.GetFeed))
}

type GetFeedReq struct {
	Limit     int64 `form:"limit"`
	Timestamp int64 `form:"timestamp"` // 用于分页，获取此时间戳之前的数据
}

type FeedArticleVO struct {
	Id         int64    `json:"id"`
	Title      string   `json:"title"`
	Abstract   string   `json:"abstract"`
	AuthorId   int64    `json:"author_id"`
	AuthorName string   `json:"author_name"`
	Tags       []string `json:"tags"`
	LikeCnt    int64    `json:"like_cnt"`
	CommentCnt int64    `json:"comment_cnt"`
	ReadCnt    int64    `json:"read_cnt"`
	Ctime      string   `json:"ctime"`
}

type FeedEventVO struct {
	Id      int64          `json:"id"`
	UserId  int64          `json:"user_id"`
	Type    string         `json:"type"`
	Content string         `json:"content"`
	Article *FeedArticleVO `json:"article,omitempty"`
	Ctime   int64          `json:"ctime"`
}

// GetFeed 获取用户 Feed 流
func (h *FeedHandler) GetFeed(ctx *gin.Context, req GetFeedReq, uc ijwt.UserClaims) (ginx.Result, error) {
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	resp, err := h.svc.FindFeedEvents(ctx.Request.Context(), &feedv1.FindFeedEventsRequest{
		Uid:       uc.UserId,
		Limit:     limit,
		Timestamp: req.Timestamp,
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
	}

	events := make([]FeedEventVO, 0, len(resp.GetFeedEvents()))
	for _, e := range resp.GetFeedEvents() {
		vo := FeedEventVO{
			Id:      e.Id,
			UserId:  e.GetUser().GetId(),
			Type:    e.Type,
			Content: e.Content,
			Ctime:   e.Ctime,
		}

		// 解析 content 获取 article_id，enrich 文章数据
		vo.Article = h.enrichFeedArticle(ctx, e.Content)

		events = append(events, vo)
	}

	return ginx.Result{
		Code: 2,
		Msg:  "OK",
		Data: events,
	}, nil
}

// enrichFeedArticle 从 feed event content 中解析 article_id 并获取文章详情
func (h *FeedHandler) enrichFeedArticle(ctx *gin.Context, content string) *FeedArticleVO {
	// content 是 JSON 序列化的 map[string]string，包含 "uid" 和 "aid"
	var ext map[string]string
	if err := json.Unmarshal([]byte(content), &ext); err != nil {
		return nil
	}
	aidStr, ok := ext["aid"]
	if !ok {
		return nil
	}
	aid, err := strconv.ParseInt(aidStr, 10, 64)
	if err != nil {
		return nil
	}

	// 获取文章详情
	artResp, err := h.articleSvc.GetPublishedById(ctx.Request.Context(), &articlev1.GetPublishedByIdRequest{
		Id: aid,
	})
	if err != nil {
		h.l.Error("feed enrichment: 获取文章详情失败", logger.Int64("aid", aid), logger.Error(err))
		return nil
	}
	art := artResp.GetArticle()
	if art == nil {
		return nil
	}

	vo := &FeedArticleVO{
		Id:       art.GetId(),
		Title:    art.GetTitle(),
		AuthorId: art.GetAuthor().GetId(),
		Ctime:    art.GetCtime().AsTime().Format(time.DateTime),
	}

	// 摘要：截取 content 前 200 字
	artContent := art.GetContent()
	if utf8.RuneCountInString(artContent) > 200 {
		vo.Abstract = string([]rune(artContent)[:200]) + "..."
	} else {
		vo.Abstract = artContent
	}

	// 获取作者名
	userResp, er := h.userSvc.Profile(ctx.Request.Context(), &userv1.ProfileRequest{Id: art.GetAuthor().GetId()})
	if er == nil && userResp.GetUser() != nil {
		vo.AuthorName = userResp.GetUser().GetNickName()
	}

	// 获取互动数据
	intrResp, er := h.intrSvc.Get(ctx.Request.Context(), &intrv1.GetRequest{
		Biz:   "article",
		BizId: aid,
	})
	if er == nil {
		vo.LikeCnt = intrResp.GetIntr().GetLikeCnt()
		vo.ReadCnt = intrResp.GetIntr().GetReadCnt()
	}

	// 获取评论数
	commentResp, er := h.commentSvc.GetCommentCnt(ctx.Request.Context(), &commentv1.GetCommentCntRequest{
		Biz:   "article",
		Bizid: aid,
	})
	if er == nil {
		vo.CommentCnt = commentResp.GetCnt()
	}

	// 获取标签
	tagResp, er := h.tagSvc.GetBizTags(ctx.Request.Context(), &tagv1.GetBizTagsRequest{
		Biz:   "article",
		BizId: aid,
	})
	if er == nil {
		tags := make([]string, 0, len(tagResp.GetTags()))
		for _, t := range tagResp.GetTags() {
			tags = append(tags, t.GetName())
		}
		vo.Tags = tags
	}
	if vo.Tags == nil {
		vo.Tags = []string{}
	}

	return vo
}
