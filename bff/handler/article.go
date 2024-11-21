package handler

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strconv"
	"time"
	articlev1 "webook/api/proto/gen/article/v1"
	intrv1 "webook/api/proto/gen/intr/v1"
	rewardv1 "webook/api/proto/gen/reward/v1"
	"webook/article/domain"
	ijwt "webook/internal/handler/jwt"
	"webook/pkg/ginx"
	"webook/pkg/logger"
)

//type ArticleHandlerV2 struct {
//	// 组合 ArticleHandlerV1，如果有变动直接重写方法
//	ArticleHandlerV1
//}
//
//func (h *ArticleHandlerV2) RegisterRoutes(server *gin.Engine) {
//	v1 := server.Group("/v1")
//	g := v1.Group("/articles")
//	g.POST("/edit", ginx.WrapBodyAndToken(h.l, h.Edit))
//}
//
//type ArticleHandlerV1 struct {
//	// 组合 ArticleHandler，如果有变动直接重写方法
//	ArticleHandler
//}
//
//func (h *ArticleHandlerV1) RegisterRoutes(server *gin.Engine) {
//	v1 := server.Group("/v1")
//	g := v1.Group("/articles")
//	g.POST("/edit", ginx.WrapBodyAndToken(h.l, h.Edit))
//}

type ArticleHandler struct {
	svc     articlev1.ArticleServiceClient
	intrSvc intrv1.InteractiveServiceClient
	reward  rewardv1.RewardServiceClient
	l       logger.LoggerV1
	biz     string
}

func NewArticleHandler(svc articlev1.ArticleServiceClient, l logger.LoggerV1, intrSvc intrv1.InteractiveServiceClient, reward rewardv1.RewardServiceClient) *ArticleHandler {
	return &ArticleHandler{
		svc:     svc,
		intrSvc: intrSvc,
		l:       l,
		biz:     "article",
		reward:  reward,
	}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")
	g.POST("/edit", ginx.WrapBodyAndToken(h.l, h.Edit))
	g.POST("/publish", ginx.WrapBodyAndToken(h.l, h.Publish))
	g.POST("/withdraw", ginx.WrapBodyAndToken(h.l, h.Withdraw))
	// 创作者的查询接口
	g.POST("/list", ginx.WrapBodyAndToken[ListReq, ijwt.UserClaims](h.l, h.List))
	g.GET("/detail/:id", ginx.WrapClaims(h.l, h.Detail))

	pub := g.Group("/pub")
	pub.GET("/:id", ginx.WrapClaims(h.l, h.PubDetail))

	pub.POST("/like", ginx.WrapBodyAndToken(h.l, h.Like))
	pub.POST("/collect", ginx.WrapBodyAndToken[LikeReq, ijwt.UserClaims](h.l, h.Like))

	pub.POST("/reward", ginx.WrapBodyAndToken(h.l, h.Reward))
}

//type ReaderHandler struct {
//	pub := g.Group("/pub")
//	pub.GET("/:id")
//}

func (h *ArticleHandler) Like(ctx *gin.Context, req LikeReq, uc ijwt.UserClaims) (ginx.Result, error) {
	var err error
	if req.Like {
		_, err = h.intrSvc.Like(ctx.Request.Context(), &intrv1.LikeRequest{
			Biz:   h.biz,
			BizId: req.Id,
			Uid:   uc.UserId,
		})
	} else {
		_, err = h.intrSvc.CancelLike(ctx.Request.Context(), &intrv1.CancelLikeRequest{
			Biz:   h.biz,
			BizId: req.Id,
			Uid:   uc.UserId,
		})
	}

	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{Msg: "OK"}, nil
}

func (h *ArticleHandler) PubDetail(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return ginx.Result{
			Code: 4,
			Msg:  "参数错误",
		}, errors.New("前端输入的 ID 不对")
	}

	var eg errgroup.Group
	var art domain.Article
	eg.Go(func() error {
		// 读文章本体
		resp, err := h.svc.GetPublishedById(ctx.Request.Context(), &articlev1.GetPublishedByIdRequest{
			Id:  id,
			Uid: uc.UserId,
		})
		// TODO 处理空错误
		art = domain.Article{
			Id:      resp.GetArticle().GetId(),
			Title:   resp.GetArticle().GetTitle(),
			Content: resp.GetArticle().GetContent(),
			Author: domain.Author{
				Id:   resp.GetArticle().GetAuthor().GetId(),
				Name: resp.GetArticle().GetAuthor().GetName(),
			},
			Status: domain.ArticleStatus(resp.GetArticle().GetStatus()),
			Ctime:  resp.GetArticle().GetCtime().AsTime(),
			Utime:  resp.GetArticle().GetUtime().AsTime(),
		}
		return err
	})

	// 要在这里获得这篇文章的全部计数
	// 可以容忍这个错误
	var getResp *intrv1.GetResponse
	eg.Go(func() error {
		getResp, err = h.intrSvc.Get(ctx.Request.Context(), &intrv1.GetRequest{
			Biz:   h.biz,
			BizId: id,
			Uid:   uc.UserId,
		})
		if err != nil {
			// 几率日志
			h.l.Error("查询文章相关数据失败", logger.Error(err))
		}
		return nil
	})

	// 等待
	err = eg.Wait()
	if err != nil {
		// 代表查询出错了
		return ginx.Result{
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

	intr := getResp.Intr

	// 这个功能是不是可以让前端，主动发一个 HTTP 请求，来增加一个计数？
	return ginx.Result{
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

func (h *ArticleHandler) Detail(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.l.Error("前端输入的 ID 不对", logger.Error(err))
		return ginx.Result{
			Code: 4,
			Msg:  "参数错误",
		}, err
	}
	resp, err := h.svc.GetById(ctx.Request.Context(), &articlev1.GetByIdRequest{
		Id: int64(id),
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	art := resp.GetArticle()
	if art.Author.Id != uc.UserId {
		return ginx.Result{
			Code: 4,
			// 不需要告诉前端究竟发生了什么
			Msg: "输入有误",
		}, fmt.Errorf("非法访问文章，创作者 ID 不匹配 %d", uc.UserId)
	}
	return ginx.Result{
		Data: ArticleVO{
			Id:    art.Id,
			Title: art.Title,
			// 不需要摘要详细信息
			//Abstract: art.Abstract(),
			Status:  uint8(art.Status),
			Content: art.Content,
			// 这个接口是创作者看自己的文章，不需要这个字段
			//AuthorId:   art.Author.Id,
			//AuthorName: art.Author.Name,
			Ctime: art.Ctime.AsTime().Format(time.DateTime),
			Utime: art.Utime.AsTime().Format(time.DateTime),
		},
	}, nil
}

func (h *ArticleHandler) Withdraw(ctx *gin.Context, req WithdrawReq, uc ijwt.UserClaims) (ginx.Result, error) {
	_, err := h.svc.WithDraw(ctx.Request.Context(), &articlev1.WithDrawRequest{
		Article: &articlev1.Article{
			Id: req.Id,
			Author: &articlev1.Author{
				Id: uc.UserId,
			},
		},
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, fmt.Errorf("撤回帖子失败 %w", err)
	}
	return ginx.Result{Msg: "OK", Data: req.Id}, nil
}

func (h *ArticleHandler) Publish(ctx *gin.Context, req ArticleReq, uc ijwt.UserClaims) (ginx.Result, error) {
	art := req.toDomain(req, &uc)
	resp, err := h.svc.Publish(ctx.Request.Context(), &articlev1.PublishRequest{
		Article: &articlev1.Article{
			Id:      art.Id,
			Title:   art.Title,
			Content: art.Content,
			Author: &articlev1.Author{
				Id:   art.Author.Id,
				Name: art.Author.Name,
			},
			Status: int32(art.Status),
			Ctime:  timestamppb.New(art.Ctime),
			Utime:  timestamppb.New(art.Utime),
		},
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, fmt.Errorf("发表帖子失败 %w", err)
	}
	id := resp.GetId()
	return ginx.Result{Msg: "OK", Data: id}, nil
}

func (h *ArticleHandler) Edit(ctx *gin.Context, req ArticleReq, uc ijwt.UserClaims) (ginx.Result, error) {
	// TODO: 检测输入
	// 调用 service 的代码
	art := req.toDomain(req, &uc)
	id, err := h.svc.Save(ctx.Request.Context(), &articlev1.SaveRequest{Article: &articlev1.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: &articlev1.Author{
			Id:   art.Author.Id,
			Name: art.Author.Name,
		},
		Status: int32(art.Status),
		Ctime:  timestamppb.New(art.Ctime),
		Utime:  timestamppb.New(art.Utime),
	}})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, fmt.Errorf("保存帖子失败 %w", err)
	}
	return ginx.Result{Msg: "OK", Data: id}, nil
}

func (h *ArticleHandler) List(ctx *gin.Context, req ListReq, uc ijwt.UserClaims) (ginx.Result, error) {
	resp, err := h.svc.List(ctx.Request.Context(), &articlev1.ListRequest{
		Uid:    uc.UserId,
		Offset: int32(req.Offset),
		Limit:  int32(req.Limit),
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, nil
	}

	res := resp.GetArticles()
	arts := make([]domain.Article, 0, len(res))
	for i, art := range res {
		arts[i] = domain.Article{
			Id:      art.Id,
			Title:   art.Title,
			Content: art.Content,
			Author: domain.Author{
				Id:   art.Author.Id,
				Name: art.Author.Name,
			},
			Status: domain.ArticleStatus(art.Status),
			Ctime:  art.Ctime.AsTime(),
			Utime:  art.Utime.AsTime(),
		}
	}

	// 在列表页，不显示全文，只显示一个"摘要"
	// 比如说，简单的摘要就是前几句话
	// 强大的摘要是 AI 帮你生成的
	return ginx.Result{
		Data: req.articlesToVO(arts),
	}, err
}

// Reward 打赏
func (h *ArticleHandler) Reward(ctx *gin.Context, req RewardReq, uc ijwt.UserClaims) (ginx.Result, error) {
	artResp, err := h.svc.GetPublishedById(ctx.Request.Context(), &articlev1.GetPublishedByIdRequest{
		Id: req.Id,
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	art := artResp.GetArticle()
	resp, err := h.reward.PreReward(ctx.Request.Context(), &rewardv1.PreRewardRequest{
		Biz:       "article",
		BizId:     art.Id,
		BizName:   art.Title,
		TargetUid: art.Author.Id,
		Uid:       uc.UserId,
		Amt:       req.Amt,
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Data: map[string]interface{}{
			"codeURL": resp.CodeUrl,
			"rid":     resp.Rid,
		},
	}, nil
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
