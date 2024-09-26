package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"webook/internal/domain"
	ijwt "webook/internal/handler/jwt"
	"webook/internal/pkg/logger"
	"webook/internal/service"
)

//var _ handler = (*ArticleHandler)(nil)

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
	// 新增
	//g.POST("/")
	// 修改
	//g.PUT("/")

	g.POST("/edit", h.Edit)
	g.POST("/publish", h.Publish)
	g.POST("/withdraw", h.Withdraw)
}

func (h *ArticleHandler) Withdraw(ctx *gin.Context) {
	type Req struct {
		Id int64
	}
	var req Req
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "参数错误",
		})
		return
	}
	c, _ := ctx.Get("claims")
	claims, ok := c.(*ijwt.UserClaims)
	if !ok {
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("未发现用户的 session 信息")
		return
	}
	err := h.svc.WithDraw(ctx, domain.Article{
		Id: req.Id,
		Author: domain.Author{
			Id: claims.UserId,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("撤回帖子失败", logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg:  "OK",
		Data: req.Id,
	})

}

func (h *ArticleHandler) Publish(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "参数错误",
		})
		return
	}
	c, _ := ctx.Get("claims")
	claims, ok := c.(*ijwt.UserClaims)
	if !ok {
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("未发现用户的 session 信息")
		return
	}
	id, err := h.svc.Publish(ctx, req.toDomain(req, claims))
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("发表帖子失败", logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg:  "OK",
		Data: id,
	})
}

func (h *ArticleHandler) Edit(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "参数错误",
		})
		return
	}
	c, _ := ctx.Get("claims")
	claims, ok := c.(*ijwt.UserClaims)
	if !ok {
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("未发现用户的 session 信息")
		return
	}
	// TODO: 检测输入
	// 调用 service 的代码
	id, err := h.svc.Save(ctx, req.toDomain(req, claims))
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("保存帖子失败", logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg:  "OK",
		Data: id,
	})
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
