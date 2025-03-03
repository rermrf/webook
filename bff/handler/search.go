package handler

import (
	"github.com/gin-gonic/gin"
	searchv1 "webook/api/proto/gen/search/v1"
	ijwt "webook/bff/handler/jwt"
	"webook/pkg/ginx"
	"webook/pkg/logger"
)

type SearchHandler struct {
	svc searchv1.SearchServiceClient
	l   logger.LoggerV1
}

func NewSearchHandler(l logger.LoggerV1, svc searchv1.SearchServiceClient) *SearchHandler {
	return &SearchHandler{l: l, svc: svc}
}

func (h *SearchHandler) RegisterRoutes(server *gin.Engine) {
	server.GET("/search", ginx.WrapBodyAndToken(h.l, h.Search))
}

type SearchRequest struct {
	Expression string `form:"expression"`
}

type User struct {
	Id       int64  `json:"id"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

type Article struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Status  uint8  `json:"status"`
	Content string `json:"content"`
}

type SearchResponse struct {
	Users    []User    `json:"users"`
	Articles []Article `json:"articles"`
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
		Users:    []User{},
		Articles: []Article{},
	}
	for _, user := range resp.GetUser().GetUsers() {
		res.Users = append(res.Users, User{
			Id:       user.Id,
			Nickname: user.Nickname,
			Email:    user.Email,
			Phone:    user.Phone,
		})
	}
	for _, article := range resp.GetArticle().GetArticles() {
		res.Articles = append(res.Articles, Article{
			Id:      article.Id,
			Title:   article.Title,
			Content: article.Content,
			Status:  uint8(article.Status),
		})
	}
	return ginx.Result{
		Code: 2,
		Msg:  "OK",
		Data: res,
	}, nil
}
