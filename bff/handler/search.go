package handler

import (
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	followv1 "webook/api/proto/gen/follow/v1"
	intrv1 "webook/api/proto/gen/intr/v1"
	rankingv1 "webook/api/proto/gen/ranking/v1"
	searchv1 "webook/api/proto/gen/search/v1"
	userv1 "webook/api/proto/gen/user/v1"
	ijwt "webook/bff/handler/jwt"
	"webook/pkg/ginx"
	"webook/pkg/logger"
)

type SearchHandler struct {
	svc        searchv1.SearchServiceClient
	userSvc    userv1.UserServiceClient
	intrSvc    intrv1.InteractiveServiceClient
	followSvc  followv1.FollowServiceClient
	rankingSvc rankingv1.RankingServiceClient
	l          logger.LoggerV1
}

func NewSearchHandler(l logger.LoggerV1, svc searchv1.SearchServiceClient, userSvc userv1.UserServiceClient, intrSvc intrv1.InteractiveServiceClient, followSvc followv1.FollowServiceClient, rankingSvc rankingv1.RankingServiceClient) *SearchHandler {
	return &SearchHandler{l: l, svc: svc, userSvc: userSvc, intrSvc: intrSvc, followSvc: followSvc, rankingSvc: rankingSvc}
}

func (h *SearchHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/search")
	g.GET("", ginx.WrapBodyAndToken(h.l, h.Search))
	g.GET("/hot-keywords", h.HotKeywords)
}

type SearchRequest struct {
	Expression string `form:"expression"`
}

type SearchUser struct {
	Id            int64  `json:"id"`
	Nickname      string `json:"nickname"`
	AboutMe       string `json:"about_me"`
	FollowerCount int64  `json:"follower_count"`
}

type SearchArticle struct {
	Id         int64    `json:"id"`
	Title      string   `json:"title"`
	Abstract   string   `json:"abstract"`
	AuthorId   int64    `json:"author_id"`
	AuthorName string   `json:"author_name"`
	Tags       []string `json:"tags"`
	LikeCnt    int64    `json:"like_cnt"`
	Ctime      string   `json:"ctime"`
}

type SearchResponse struct {
	Users    []SearchUser    `json:"users"`
	Articles []SearchArticle `json:"articles"`
}

func (h *SearchHandler) Search(ctx *gin.Context, req SearchRequest, uc ijwt.UserClaims) (ginx.Result, error) {
	resp, err := h.svc.Search(ctx, &searchv1.SearchRequest{
		Expression: req.Expression,
		Uid:        uc.UserId,
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, nil
	}
	res := SearchResponse{
		Users:    []SearchUser{},
		Articles: []SearchArticle{},
	}

	// 处理用户结果：enrichment 粉丝数和 about_me
	for _, user := range resp.GetUser().GetUsers() {
		su := SearchUser{
			Id:       user.Id,
			Nickname: user.Nickname,
		}
		// 获取用户详细信息
		profileResp, er := h.userSvc.Profile(ctx.Request.Context(), &userv1.ProfileRequest{Id: user.Id})
		if er == nil && profileResp.GetUser() != nil {
			su.AboutMe = profileResp.GetUser().GetAboutMe()
			su.Nickname = profileResp.GetUser().GetNickName()
		}
		// 获取粉丝数
		staticResp, er := h.followSvc.GetFollowStatic(ctx.Request.Context(), &followv1.GetFollowStaticRequest{
			Followee: user.Id,
		})
		if er == nil {
			su.FollowerCount = staticResp.GetFollowStatic().GetFollowers()
		}
		res.Users = append(res.Users, su)
	}

	// 处理文章结果：enrichment 作者名、互动数据
	for _, article := range resp.GetArticle().GetArticles() {
		sa := SearchArticle{
			Id:    article.Id,
			Title: article.Title,
			Tags:  article.Tags,
		}
		if sa.Tags == nil {
			sa.Tags = []string{}
		}

		// 截取 content 前 200 字作为摘要
		content := article.Content
		if utf8.RuneCountInString(content) > 200 {
			runes := []rune(content)
			sa.Abstract = string(runes[:200]) + "..."
		} else {
			sa.Abstract = content
		}

		// 获取互动数据
		intrResp, er := h.intrSvc.Get(ctx.Request.Context(), &intrv1.GetRequest{
			Biz:   "article",
			BizId: article.Id,
		})
		if er == nil {
			sa.LikeCnt = intrResp.GetIntr().GetLikeCnt()
		}

		// 获取作者信息 — 搜索 proto 中的 Article 没有 author_id，
		// 需要通过其他方式获取。这里跳过 author enrichment，因为搜索索引中没有 author_id。
		// 如果后续搜索 proto 扩展了 author 字段，可以在此处补充。

		res.Articles = append(res.Articles, sa)
	}

	return ginx.Result{
		Code: 2,
		Msg:  "OK",
		Data: res,
	}, nil
}

// HotKeywords 获取热门搜索关键词（基于热榜文章标题提取）
func (h *SearchHandler) HotKeywords(ctx *gin.Context) {
	type HotKeyword struct {
		Keyword string `json:"keyword"`
	}

	resp, err := h.rankingSvc.TopN(ctx.Request.Context(), &rankingv1.TopNRequest{})
	if err != nil {
		ctx.JSON(200, ginx.Result{Code: 5, Msg: "系统错误"})
		return
	}

	keywords := make([]HotKeyword, 0, len(resp.GetArticles()))
	for _, art := range resp.GetArticles() {
		keywords = append(keywords, HotKeyword{
			Keyword: art.GetTitle(),
		})
	}

	ctx.JSON(200, ginx.Result{
		Code: 2,
		Msg:  "OK",
		Data: keywords,
	})
}
