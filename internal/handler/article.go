package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"webook/internal/domain"
	ijwt "webook/internal/handler/jwt"
	"webook/internal/pkg/gin-pulgin"
	"webook/internal/pkg/logger"
	"webook/internal/service"
)

type ArticleHandler struct {
	svc service.ArticleService
	l   logger.LoggerV1
}

func NewArticleHandler(svc service.ArticleService, l logger.LoggerV1) *ArticleHandler {
	return &ArticleHandler{
		svc: svc,
		l:   l,
	}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")
	g.POST("/edit", gin_pulgin.WrapBodyAndToken(h.l, h.Edit))
	g.POST("/publish", gin_pulgin.WrapBodyAndToken(h.l, h.Publish))
	g.POST("/withdraw", gin_pulgin.WrapBodyAndToken(h.l, h.Withdraw))
}

type WithdrawReq struct {
	Id int64
}

func (h *ArticleHandler) Withdraw(ctx *gin.Context, req WithdrawReq, uc ijwt.UserClaims) (gin_pulgin.Result, error) {
	err := h.svc.WithDraw(ctx, domain.Article{
		Id: req.Id,
		Author: domain.Author{
			Id: uc.UserId,
		},
	})
	if err != nil {
		return gin_pulgin.Result{Code: 5, Msg: "系统错误"}, fmt.Errorf("撤回帖子失败 %w", err)
	}
	return gin_pulgin.Result{Msg: "OK", Data: req.Id}, nil
}

func (h *ArticleHandler) Publish(ctx *gin.Context, req ArticleReq, uc ijwt.UserClaims) (gin_pulgin.Result, error) {
	id, err := h.svc.Publish(ctx, req.toDomain(req, &uc))
	if err != nil {
		return gin_pulgin.Result{Code: 5, Msg: "系统错误"}, fmt.Errorf("发表帖子失败 %w", err)
	}
	return gin_pulgin.Result{Msg: "OK", Data: id}, nil
}

func (h *ArticleHandler) Edit(ctx *gin.Context, req ArticleReq, uc ijwt.UserClaims) (gin_pulgin.Result, error) {
	// TODO: 检测输入
	// 调用 service 的代码
	id, err := h.svc.Save(ctx, req.toDomain(req, &uc))
	if err != nil {
		ctx.JSON(http.StatusOK, gin_pulgin.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return gin_pulgin.Result{Code: 5, Msg: "系统错误"}, fmt.Errorf("保存帖子失败 %w", err)
	}
	return gin_pulgin.Result{Msg: "OK", Data: id}, nil
}

type ArticleReq struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (r ArticleReq) toDomain(req ArticleReq, claims *ijwt.UserClaims) domain.Article {
	return domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author:  domain.Author{Id: claims.UserId},
	}
}
