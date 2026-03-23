package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	commentv1 "webook/api/proto/gen/comment/v1"
	intrv1 "webook/api/proto/gen/intr/v1"
	rankingv1 "webook/api/proto/gen/ranking/v1"
	"webook/pkg/ginx"
	"webook/pkg/logger"
)

type RankingHandler struct {
	svc        rankingv1.RankingServiceClient
	intrSvc    intrv1.InteractiveServiceClient
	commentSvc commentv1.CommentServiceClient
	l          logger.LoggerV1
}

func NewRankingHandler(svc rankingv1.RankingServiceClient, intrSvc intrv1.InteractiveServiceClient, commentSvc commentv1.CommentServiceClient, l logger.LoggerV1) *RankingHandler {
	return &RankingHandler{svc: svc, intrSvc: intrSvc, commentSvc: commentSvc, l: l}
}

func (h *RankingHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/ranking")
	g.GET("/hot", h.GetHotArticles) // 获取热门文章
}

type HotArticleVO struct {
	Id         int64  `json:"id"`
	Title      string `json:"title"`
	AuthorId   int64  `json:"author_id"`
	AuthorName string `json:"author_name"`
	LikeCnt    int64  `json:"like_cnt"`
	CommentCnt int64  `json:"comment_cnt"`
	ReadCnt    int64  `json:"read_cnt"`
	Ctime      string `json:"ctime"`
}

// GetHotArticles 获取热门文章榜单
func (h *RankingHandler) GetHotArticles(ctx *gin.Context) {
	resp, err := h.svc.TopN(ctx.Request.Context(), &rankingv1.TopNRequest{})
	if err != nil {
		ctx.JSON(200, ginx.Result{Code: 5, Msg: "系统错误"})
		return
	}

	articles := make([]HotArticleVO, 0, len(resp.GetArticles()))
	for _, art := range resp.GetArticles() {
		vo := HotArticleVO{
			Id:         art.Id,
			Title:      art.Title,
			AuthorId:   art.GetAuthor().GetId(),
			AuthorName: art.GetAuthor().GetName(),
			Ctime:      art.Ctime.AsTime().Format(time.DateTime),
		}

		// 获取互动数据
		intrResp, er := h.intrSvc.Get(ctx.Request.Context(), &intrv1.GetRequest{
			Biz:   "article",
			BizId: art.Id,
		})
		if er == nil {
			vo.LikeCnt = intrResp.GetIntr().GetLikeCnt()
			vo.ReadCnt = intrResp.GetIntr().GetReadCnt()
		}

		// 获取评论数
		commentResp, er := h.commentSvc.GetCommentCnt(ctx.Request.Context(), &commentv1.GetCommentCntRequest{
			Biz:   "article",
			Bizid: art.Id,
		})
		if er == nil {
			vo.CommentCnt = commentResp.GetCnt()
		}

		articles = append(articles, vo)
	}

	ctx.JSON(200, ginx.Result{
		Code: 2,
		Msg:  "OK",
		Data: articles,
	})
}
