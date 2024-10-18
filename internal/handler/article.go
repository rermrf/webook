package handler

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"strconv"
	"time"
	"webook/internal/domain"
	ijwt "webook/internal/handler/jwt"
	"webook/internal/pkg/gin-pulgin"
	"webook/internal/pkg/logger"
	"webook/internal/service"
)

type ArticleHandler struct {
	svc     service.ArticleService
	intrSvc service.InteractiveService
	l       logger.LoggerV1
	biz     string
}

func NewArticleHandler(svc service.ArticleService, l logger.LoggerV1, intrSvc service.InteractiveService) *ArticleHandler {
	return &ArticleHandler{
		svc:     svc,
		intrSvc: intrSvc,
		l:       l,
		biz:     "article",
	}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")
	g.POST("/edit", gin_pulgin.WrapBodyAndToken(h.l, h.Edit))
	g.POST("/publish", gin_pulgin.WrapBodyAndToken(h.l, h.Publish))
	g.POST("/withdraw", gin_pulgin.WrapBodyAndToken(h.l, h.Withdraw))
	// 创作者的查询接口
	g.POST("/list", gin_pulgin.WrapBodyAndToken[ListReq, ijwt.UserClaims](h.l, h.List))
	g.GET("/detail/:id", gin_pulgin.WrapClaims(h.l, h.Detail))

	pub := g.Group("/pub")
	pub.GET("/:id", gin_pulgin.WrapClaims(h.l, h.PubDetail))

	pub.POST("/like", gin_pulgin.WrapBodyAndToken(h.l, h.Like))
	pub.POST("/collect", gin_pulgin.WrapBodyAndToken[LikeReq, ijwt.UserClaims](h.l, h.Like))
}

//type ReaderHandler struct {
//	pub := g.Group("/pub")
//	pub.GET("/:id")
//}

func (h *ArticleHandler) Like(ctx *gin.Context, req LikeReq, uc ijwt.UserClaims) (gin_pulgin.Result, error) {
	var err error
	if req.Like {
		err = h.intrSvc.Like(ctx, h.biz, req.Id, uc.UserId)
	} else {
		err = h.intrSvc.CancelLike(ctx, h.biz, req.Id, uc.UserId)
	}

	if err != nil {
		return gin_pulgin.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return gin_pulgin.Result{Msg: "OK"}, nil
}

func (h *ArticleHandler) PubDetail(ctx *gin.Context, uc ijwt.UserClaims) (gin_pulgin.Result, error) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return gin_pulgin.Result{
			Code: 4,
			Msg:  "参数错误",
		}, errors.New("前端输入的 ID 不对")
	}

	var eg errgroup.Group
	var art domain.Article
	eg.Go(func() error {
		// 读文章本体
		art, err = h.svc.GetPublishedById(ctx, id, uc.UserId)
		return err
	})

	// 要在这里获得这篇文章的全部计数
	// 可以容忍这个错误
	var intr domain.Interactive
	eg.Go(func() error {
		intr, err = h.intrSvc.Get(ctx, h.biz, id, uc.UserId)
		if err != nil {
			// 几率日志
			h.l.Error("查询文章相关数据失败", logger.Error(err))
		}
		return nil
		//return err
	})

	// 等待
	err = eg.Wait()
	if err != nil {
		// 代表查询出错了
		return gin_pulgin.Result{
			Code: 5,
			Msg:  "系统错误",
		}, errors.New("获得文章信息失败")
	}

	// 增加阅读计数
	//go func() {
	//	er := h.intrSvc.IncrReadCnt(ctx, h.biz, art.Id)
	//	if er != nil {
	//		h.l.Error("增加阅读计数失败", logger.Int64("artId", art.Id), logger.Error(er))
	//	}
	//}()

	// 这个功能是不是可以让前端，主动发一个 HTTP 请求，来增加一个计数？
	return gin_pulgin.Result{
		Data: ArticleVO{
			Id:      art.Id,
			Title:   art.Title,
			Status:  art.Status.ToUint8(),
			Content: art.Content,
			// 要把作者信息带出去
			AuthorId:   art.Author.Id,
			AuthorName: art.Author.Name,
			Ctime:      art.Ctime.Format(time.DateTime),
			Utime:      art.Utime.Format(time.DateTime),
			Liked:      intr.Liked,
			Collected:  intr.Collected,
			ReadCnt:    intr.ReadCnt,
			LikeCnt:    intr.LikeCnt,
			CollectCnt: intr.CollectCnt,
		},
	}, nil
}

func (h *ArticleHandler) Detail(ctx *gin.Context, uc ijwt.UserClaims) (gin_pulgin.Result, error) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.l.Error("前端输入的 ID 不对", logger.Error(err))
		return gin_pulgin.Result{
			Code: 4,
			Msg:  "参数错误",
		}, err
	}
	art, err := h.svc.GetById(ctx, int64(id))
	if err != nil {
		return gin_pulgin.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	if art.Author.Id != uc.UserId {
		return gin_pulgin.Result{
			Code: 4,
			// 不需要告诉前端究竟发生了什么
			Msg: "输入有误",
		}, fmt.Errorf("非法访问文章，创作者 ID 不匹配 %d", uc.UserId)
	}
	return gin_pulgin.Result{
		Data: ArticleVO{
			Id:    art.Id,
			Title: art.Title,
			// 不需要摘要详细信息
			//Abstract: art.Abstract(),
			Status:  art.Status.ToUint8(),
			Content: art.Content,
			// 这个接口是创作者看自己的文章，不需要这个字段
			//AuthorId:   art.Author.Id,
			//AuthorName: art.Author.Name,
			Ctime: art.Ctime.Format(time.DateTime),
			Utime: art.Utime.Format(time.DateTime),
		},
	}, nil
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
		return gin_pulgin.Result{Code: 5, Msg: "系统错误"}, fmt.Errorf("保存帖子失败 %w", err)
	}
	return gin_pulgin.Result{Msg: "OK", Data: id}, nil
}

func (h *ArticleHandler) List(ctx *gin.Context, req ListReq, uc ijwt.UserClaims) (gin_pulgin.Result, error) {
	res, err := h.svc.List(ctx, uc.UserId, req.Offset, req.Limit)
	if err != nil {
		return gin_pulgin.Result{Code: 5, Msg: "系统错误"}, nil
	}
	// 在列表页，不显示全文，只显示一个"摘要"
	// 比如说，简单的摘要就是前几句话
	// 强大的摘要是 AI 帮你生成的
	return gin_pulgin.Result{
		Data: req.articlesToVO(res),
	}, err
}

func (r ListReq) articlesToVO(src []domain.Article) []ArticleVO {
	var res []ArticleVO
	for _, v := range src {
		res = append(res, ArticleVO{
			Id:       v.Id,
			Title:    v.Title,
			Abstract: v.Abstract(),
			Status:   v.Status.ToUint8(),
			// 这个列表请求，不需要返回内容
			//Content:  v.Content,
			// 这个是创作者看自己的文章列表，也不需要这个字段
			//AuthorId: v.Author.Id,
			Ctime: v.Ctime.Format(time.DateTime),
			Utime: v.Utime.Format(time.DateTime),
		})
	}
	return res
}

func (r ArticleReq) toDomain(req ArticleReq, claims *ijwt.UserClaims) domain.Article {
	return domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author:  domain.Author{Id: claims.UserId},
	}
}
