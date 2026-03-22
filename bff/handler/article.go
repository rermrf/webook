package handler

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"
	articlev1 "webook/api/proto/gen/article/v1"
	commentv1 "webook/api/proto/gen/comment/v1"
	followv1 "webook/api/proto/gen/follow/v1"
	historyv1 "webook/api/proto/gen/history/v1"
	intrv1 "webook/api/proto/gen/intr/v1"
	rewardv1 "webook/api/proto/gen/reward/v1"
	tagv1 "webook/api/proto/gen/tag/v1"
	userv1 "webook/api/proto/gen/user/v1"
	"webook/article/domain"
	"webook/bff/events"
	ijwt "webook/bff/handler/jwt"
	"webook/pkg/ginx"
	"webook/pkg/logger"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"
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
	svc            articlev1.ArticleServiceClient
	intrSvc        intrv1.InteractiveServiceClient
	reward         rewardv1.RewardServiceClient
	commentSvc     commentv1.CommentServiceClient
	userSvc        userv1.UserServiceClient
	followSvc      followv1.FollowServiceClient
	tagSvc         tagv1.TagServiceClient
	historySvc     historyv1.HistoryServiceClient
	notifyProducer events.NotificationProducer
	l              logger.LoggerV1
	biz            string
}

func NewArticleHandler(svc articlev1.ArticleServiceClient, l logger.LoggerV1, intrSvc intrv1.InteractiveServiceClient, reward rewardv1.RewardServiceClient, commentSvc commentv1.CommentServiceClient, userSvc userv1.UserServiceClient, followSvc followv1.FollowServiceClient, tagSvc tagv1.TagServiceClient, historySvc historyv1.HistoryServiceClient, notifyProducer events.NotificationProducer) *ArticleHandler {
	return &ArticleHandler{
		svc:            svc,
		intrSvc:        intrSvc,
		commentSvc:     commentSvc,
		userSvc:        userSvc,
		followSvc:      followSvc,
		tagSvc:         tagSvc,
		historySvc:     historySvc,
		notifyProducer: notifyProducer,
		l:              l,
		biz:            "article",
		reward:         reward,
	}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")
	g.POST("/edit", ginx.WrapBodyAndToken(h.l, h.Edit))
	g.POST("/publish", ginx.WrapBodyAndToken(h.l, h.Publish))
	g.POST("/withdraw", ginx.WrapBodyAndToken(h.l, h.Withdraw))
	g.DELETE("/:id", ginx.WrapClaims(h.l, h.Delete)) // 删除文章
	// 创作者的查询接口
	g.POST("/list", ginx.WrapBodyAndToken[ListReq, ijwt.UserClaims](h.l, h.List))
	g.GET("/detail/:id", ginx.WrapClaims(h.l, h.Detail))

	pub := g.Group("/pub")
	pub.GET("/articles", ginx.WrapBodyAndToken(h.l, h.PubList))
	pub.GET("/:id", ginx.WrapClaims(h.l, h.PubDetail))

	pub.POST("/like", ginx.WrapBodyAndToken(h.l, h.Like))
	pub.POST("/collect", ginx.WrapBodyAndToken[CollectReq, ijwt.UserClaims](h.l, h.Collect))

	pub.POST("/reward", ginx.WrapBodyAndToken(h.l, h.Reward))
	// 获取交互数据
	pub.GET("/interactive", ginx.WrapBodyAndToken(h.l, h.GetInteractive))
	// 获取评论数据
	pub.GET("/comment", ginx.WrapBody(h.l, h.GetComment))
	pub.GET("/comment_cnt", ginx.WrapBody(h.l, h.GetCommentCnt))
	// 添加评论，传入父 parent 的id，那么就是代表回复了某个评论
	pub.POST("/comment", ginx.WrapBodyAndToken(h.l, h.CreateComment))
	pub.DELETE("/comment/:id", ginx.WrapClaims(h.l, h.DeleteComment))       // 删除评论
	pub.GET("/comment/:id/replies", ginx.WrapBody(h.l, h.GetMoreReplies)) // 获取更多回复

	// 用户点赞/收藏列表
	pub.GET("/liked", ginx.WrapClaims(h.l, h.ListLiked))
	pub.GET("/collected", ginx.WrapClaims(h.l, h.ListCollected))
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
		if err == nil {
			// 发送点赞通知（异步，不阻塞响应）
			go func() {
				// 使用新的 context，避免 HTTP 请求结束后 context 被取消
				newCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()
				// 获取文章信息
				artResp, er := h.svc.GetPublishedById(newCtx, &articlev1.GetPublishedByIdRequest{
					Id: req.Id,
				})
				if er != nil {
					h.l.Error("获取文章信息失败", logger.Error(er))
					return
				}
				// 不通知自己
				if artResp.GetArticle().GetAuthor().GetId() == uc.UserId {
					return
				}
				// 获取点赞者名称
				userResp, er := h.userSvc.Profile(newCtx, &userv1.ProfileRequest{Id: uc.UserId})
				userName := "用户"
				if er == nil {
					userName = userResp.GetUser().GetNickName()
				}
				er = h.notifyProducer.ProduceLikeEvent(newCtx, events.LikeEvent{
					Uid:        uc.UserId,
					UserName:   userName,
					Biz:        h.biz,
					BizId:      req.Id,
					BizTitle:   artResp.GetArticle().GetTitle(),
					BizOwnerId: artResp.GetArticle().GetAuthor().GetId(),
				})
				if er != nil {
					h.l.Error("发送点赞通知失败", logger.Error(er))
				}
			}()
		}
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

func (h *ArticleHandler) Collect(ctx *gin.Context, req CollectReq, uc ijwt.UserClaims) (ginx.Result, error) {
	var err error
	if req.Collect {
		_, err = h.intrSvc.Collect(ctx.Request.Context(), &intrv1.CollectRequest{
			Biz:   h.biz,
			BizId: req.Id,
			Uid:   uc.UserId,
		})
		if err == nil {
			// 发送收藏通知（异步）
			go func() {
				// 使用新的 context，避免 HTTP 请求结束后 context 被取消
				newCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()
				artResp, er := h.svc.GetPublishedById(newCtx, &articlev1.GetPublishedByIdRequest{
					Id: req.Id,
				})
				if er != nil {
					h.l.Error("获取文章信息失败", logger.Error(er))
					return
				}
				// 不通知自己
				if artResp.GetArticle().GetAuthor().GetId() == uc.UserId {
					return
				}
				userResp, er := h.userSvc.Profile(newCtx, &userv1.ProfileRequest{Id: uc.UserId})
				userName := "用户"
				if er == nil {
					userName = userResp.GetUser().GetNickName()
				}
				er = h.notifyProducer.ProduceCollectEvent(newCtx, events.CollectEvent{
					Uid:        uc.UserId,
					UserName:   userName,
					Biz:        h.biz,
					BizId:      req.Id,
					BizTitle:   artResp.GetArticle().GetTitle(),
					BizOwnerId: artResp.GetArticle().GetAuthor().GetId(),
				})
				if er != nil {
					h.l.Error("发送收藏通知失败", logger.Error(er))
				}
			}()
		}
	} else {
		_, err = h.intrSvc.CancelCollect(ctx.Request.Context(), &intrv1.CancelCollectRequest{
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
			// 记录日志
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
	// 读取作者信息
	var authorName string
	var authorAvatarUrl string
	userResp, err := h.userSvc.Profile(ctx.Request.Context(), &userv1.ProfileRequest{
		Id: art.Author.Id,
	})
	if err != nil {
		h.l.Error("查询文章作者相关数据失败", logger.Error(err))
	} else if userResp != nil && userResp.GetUser() != nil {
		authorName = userResp.GetUser().NickName
		authorAvatarUrl = userResp.GetUser().GetAvatarUrl()
	}

	// 检查当前用户是否关注了作者
	var authorFollowed bool
	followResp, er := h.followSvc.FollowInfo(ctx.Request.Context(), &followv1.FollowInfoRequest{
		Follower: uc.UserId,
		Followee: art.Author.Id,
	})
	if er == nil && followResp.GetFollowRelation() != nil && followResp.GetFollowRelation().GetId() > 0 {
		authorFollowed = true
	}

	// 异步记录浏览历史
	go func() {
		ctx2, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_, _ = h.historySvc.Record(ctx2, &historyv1.RecordRequest{
			UserId:     uc.UserId,
			Biz:        "article",
			BizId:      art.Id,
			BizTitle:   art.Title,
			AuthorName: authorName,
		})
	}()

	// 这个功能是不是可以让前端，主动发一个 HTTP 请求，来增加一个计数？
	return ginx.Result{
		Data: ArticleVO{
			Id:      art.Id,
			Title:   art.Title,
			Status:  art.Status.ToUint8(),
			Content: art.Content,
			// 要把作者信息带出去
			AuthorId:        art.Author.Id,
			AuthorName:      authorName,
			AuthorAvatarUrl: authorAvatarUrl,
			AuthorFollowed:  authorFollowed,
			Ctime:          art.Ctime.Format(time.DateTime),
			Utime:          art.Utime.Format(time.DateTime),
			Liked:          intr.Liked,
			Collected:      intr.Collected,
			ReadCnt:        intr.ReadCnt,
			LikeCnt:        intr.LikeCnt,
			CollectCnt:     intr.CollectCnt,
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

	// 发布成功后，如果携带了标签，则绑定标签
	if len(req.Tags) > 0 {
		_, er := h.tagSvc.AttachTags(ctx.Request.Context(), &tagv1.AttachTagsRequest{
			Tids:  req.Tags,
			Biz:   h.biz,
			BizId: id,
		})
		if er != nil {
			h.l.Error("绑定文章标签失败", logger.Int64("artId", id), logger.Error(er))
		}
	}

	return ginx.Result{Msg: "OK", Data: id}, nil
}

func (h *ArticleHandler) Edit(ctx *gin.Context, req ArticleReq, uc ijwt.UserClaims) (ginx.Result, error) {
	// id不为0说明为更新文章
	if req.Id != 0 {
		artResp, err := h.svc.GetById(ctx.Request.Context(), &articlev1.GetByIdRequest{
			Id: req.Id,
		})
		if err != nil {
			return ginx.Result{Code: 5, Msg: "该帖子并不存在"}, nil
		}
		if artResp.GetArticle().GetAuthor().GetId() != uc.UserId {
			return ginx.Result{Code: 5, Msg: "你不能修改非自己的文章"}, nil
		}
	}
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
	// 拿到打赏的二维码
	// 不是直接调用支付，而是调用打赏
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
			// 代表的是这一次打赏的id，后续在前端可通过这个 id 验证支付
			"rid": resp.Rid,
		},
	}, nil
}

func (h *ArticleHandler) GetInteractive(ctx *gin.Context, req InteractiveReq, uc ijwt.UserClaims) (ginx.Result, error) {
	//idStr := ctx.Param("id")
	//id, err := strconv.ParseInt(idStr, 10, 64)
	resp, err := h.intrSvc.Get(ctx.Request.Context(), &intrv1.GetRequest{
		Biz:   h.biz,
		BizId: req.Id,
		Uid:   uc.UserId,
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, nil
	}
	res := InteractiveResp{
		Biz:        resp.GetIntr().GetBiz(),
		BizId:      resp.GetIntr().GetBizId(),
		ReadCnt:    resp.GetIntr().GetReadCnt(),
		LikeCnt:    resp.GetIntr().GetLikeCnt(),
		CollectCnt: resp.GetIntr().GetCollectCnt(),
		Liked:      resp.GetIntr().GetLiked(),
		Collected:  resp.GetIntr().GetCollected(),
	}
	return ginx.Result{
		Code: 2,
		Msg:  "ok",
		Data: res,
	}, nil
}

func (h *ArticleHandler) GetComment(ctx *gin.Context, req GetCommentReq) (ginx.Result, error) {
	var minId int64 = 0
	var limit int64 = 100
	if req.MinId > 0 {
		minId = req.MinId
	}
	if req.Limit > 0 {
		limit = req.Limit
	}
	resp, err := h.commentSvc.GetCommentList(ctx.Request.Context(), &commentv1.GetCommentListRequest{
		Biz:   h.biz,
		Bizid: req.Id,
		MinId: minId,
		Limit: limit,
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, nil
	}

	// 收集所有 uid 并批量查询用户名
	uidSet := make(map[int64]struct{})
	for _, c := range resp.GetComments() {
		uidSet[c.Uid] = struct{}{}
	}
	userNameMap := make(map[int64]string)
	userAvatarMap := make(map[int64]string)
	for uid := range uidSet {
		userResp, er := h.userSvc.Profile(ctx.Request.Context(), &userv1.ProfileRequest{Id: uid})
		if er == nil && userResp.GetUser() != nil {
			userNameMap[uid] = userResp.GetUser().GetNickName()
			userAvatarMap[uid] = userResp.GetUser().GetAvatarUrl()
		}
	}

	res := GetCommentResp{
		Comments: make([]Comment, len(resp.GetComments())),
	}
	for i, c := range resp.GetComments() {
		var pid int64 = 0
		var rid int64 = 0
		if c.ParentComment != nil {
			pid = c.ParentComment.Id
		}
		if c.RootComment != nil {
			rid = c.RootComment.Id
		}
		res.Comments[i] = Comment{
			Id:            c.Id,
			Content:       c.Content,
			Uid:           c.Uid,
			UserName:      userNameMap[c.Uid],
			UserAvatarUrl: userAvatarMap[c.Uid],
			ParentId:      pid,
			RootId:        rid,
			Ctime:         c.Ctime.AsTime().UnixMilli(),
		}
	}
	return ginx.Result{
		Code: 2,
		Msg:  "ok",
		Data: res,
	}, nil
}

func (h *ArticleHandler) GetCommentCnt(ctx *gin.Context, req GetCommentCntReq) (ginx.Result, error) {
	resp, err := h.commentSvc.GetCommentCnt(ctx.Request.Context(), &commentv1.GetCommentCntRequest{
		Biz:   h.biz,
		Bizid: req.Id,
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, nil
	}
	return ginx.Result{
		Code: 2,
		Msg:  "ok",
		Data: GetCommentCntResp{
			Cnt: resp.GetCnt(),
		},
	}, nil
}

func (h *ArticleHandler) CreateComment(ctx *gin.Context, req CreateCommentReq, uc ijwt.UserClaims) (ginx.Result, error) {
	var parent *commentv1.Comment = nil
	var root *commentv1.Comment = nil
	if req.ParentId != 0 {
		parent = &commentv1.Comment{
			Id: req.ParentId,
		}
	}
	if req.RootId != 0 {
		root = &commentv1.Comment{
			Id: req.RootId,
		}
	}
	_, err := h.commentSvc.CreateComment(ctx.Request.Context(), &commentv1.CreateCommentRequest{
		Comment: &commentv1.Comment{
			Uid:           uc.UserId,
			Biz:           h.biz,
			Bizid:         req.Id,
			Content:       req.Content,
			ParentComment: parent,
			RootComment:   root,
		},
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, nil
	}

	// 发送评论通知（异步）
	go func() {
		// 使用新的 context，避免 HTTP 请求结束后 context 被取消
		newCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		// 获取文章信息
		artResp, er := h.svc.GetPublishedById(newCtx, &articlev1.GetPublishedByIdRequest{
			Id: req.Id,
		})
		if er != nil {
			h.l.Error("获取文章信息失败", logger.Error(er))
			return
		}
		// 获取评论者名称
		userResp, er := h.userSvc.Profile(newCtx, &userv1.ProfileRequest{Id: uc.UserId})
		userName := "用户"
		if er == nil {
			userName = userResp.GetUser().GetNickName()
		}

		// 截取评论内容摘要
		content := req.Content
		if len(content) > 50 {
			content = content[:50] + "..."
		}

		// 如果是回复评论，尝试获取被回复人的信息
		var parentUserId int64
		if req.ParentId != 0 {
			// 获取父评论信息
			commentResp, er := h.commentSvc.GetCommentList(newCtx, &commentv1.GetCommentListRequest{
				Biz:   h.biz,
				Bizid: req.Id,
				Limit: 100,
			})
			if er == nil {
				// 在评论列表中查找父评论
				for _, c := range commentResp.GetComments() {
					if c.GetId() == req.ParentId {
						parentUserId = c.GetUid()
						break
					}
				}
			}
		}

		er = h.notifyProducer.ProduceCommentEvent(newCtx, events.CommentEvent{
			Uid:             uc.UserId,
			UserName:        userName,
			Biz:             h.biz,
			BizId:           req.Id,
			BizTitle:        artResp.GetArticle().GetTitle(),
			BizOwnerId:      artResp.GetArticle().GetAuthor().GetId(),
			Content:         content,
			ParentCommentId: req.ParentId,
			ParentUserId:    parentUserId,
		})
		if er != nil {
			h.l.Error("发送评论通知失败", logger.Error(er))
		}
	}()

	return ginx.Result{
		Code: 2,
		Msg:  "ok",
	}, nil
}

func (h *ArticleHandler) PubList(ctx *gin.Context, req GetPubListReq, uc ijwt.UserClaims) (ginx.Result, error) {
	limit := req.Limit
	if req.Limit <= 0 {
		limit = 100
	}
	resp, err := h.svc.ListPub(ctx.Request.Context(), &articlev1.ListPubRequest{
		Limit:     limit,
		Offset:    req.Offset,
		StartTime: timestamppb.New(time.Now().AddDate(0, -1, 0)),
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	var res []ArticleVO
	for _, art := range resp.GetArticles() {
		// 读取作者信息
		var authorName string
		var authorAvatarUrl string
		userResp, err := h.userSvc.Profile(ctx.Request.Context(), &userv1.ProfileRequest{
			Id: art.Author.Id,
		})
		if err != nil {
			h.l.Error("查询文章作者相关数据失败", logger.Error(err))
		} else if userResp != nil && userResp.GetUser() != nil {
			authorName = userResp.GetUser().NickName
			authorAvatarUrl = userResp.GetUser().GetAvatarUrl()
		}

		intrResp, err := h.intrSvc.Get(ctx.Request.Context(), &intrv1.GetRequest{
			Biz:   h.biz,
			BizId: art.Id,
			Uid:   uc.UserId,
		})
		if err != nil {
			// 可能没有数据
		}
		res = append(res, ArticleVO{
			Id:              art.Id,
			Title:           art.Title,
			Content:         art.Content,
			AuthorId:        art.Author.Id,
			AuthorName:      authorName,
			AuthorAvatarUrl: authorAvatarUrl,
			ReadCnt:    intrResp.GetIntr().GetReadCnt(),
			LikeCnt:    intrResp.GetIntr().GetLikeCnt(),
			CollectCnt: intrResp.GetIntr().GetCollectCnt(),
			Liked:      intrResp.GetIntr().GetLiked(),
			Collected:  intrResp.GetIntr().GetCollected(),
			Status:     uint8(art.GetStatus()),
			Ctime:      art.GetCtime().AsTime().Format(time.DateTime),
			Utime:      art.GetUtime().AsTime().Format(time.DateTime),
		})
	}
	return ginx.Result{
		Code: 2,
		Msg:  "ok",
		Data: res,
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

// Delete 删除文章（仅作者可操作）
func (h *ArticleHandler) Delete(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return ginx.Result{Code: 4, Msg: "参数错误"}, err
	}

	// 先获取文章，校验是否为作者
	artResp, err := h.svc.GetById(ctx.Request.Context(), &articlev1.GetByIdRequest{Id: id})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
	}
	if artResp.GetArticle().GetAuthor().GetId() != uc.UserId {
		return ginx.Result{Code: 4, Msg: "无权限删除此文章"}, nil
	}

	// 调用撤回接口（将文章状态设为隐藏/删除状态）
	_, err = h.svc.WithDraw(ctx.Request.Context(), &articlev1.WithDrawRequest{
		Article: &articlev1.Article{
			Id: id,
			Author: &articlev1.Author{
				Id: uc.UserId,
			},
		},
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "删除失败"}, err
	}
	return ginx.Result{Msg: "删除成功"}, nil
}

// DeleteComment 删除评论（仅评论作者可操作）
func (h *ArticleHandler) DeleteComment(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return ginx.Result{Code: 4, Msg: "参数错误"}, err
	}

	// 调用 comment 服务删除评论
	// 注意：comment 服务的 DeleteComment 应该内部校验评论所有者
	_, err = h.commentSvc.DeleteComment(ctx.Request.Context(), &commentv1.DeleteCommentRequest{
		Id: id,
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "删除评论失败"}, err
	}
	return ginx.Result{Msg: "删除成功"}, nil
}

// GetMoreReplies 获取某条评论的更多回复
func (h *ArticleHandler) GetMoreReplies(ctx *gin.Context, req GetMoreRepliesReq) (ginx.Result, error) {
	ridStr := ctx.Param("id")
	rid, err := strconv.ParseInt(ridStr, 10, 64)
	if err != nil {
		return ginx.Result{Code: 4, Msg: "参数错误"}, err
	}

	var minId int64 = 0
	var limit int64 = 20
	if req.MinId > 0 {
		minId = req.MinId
	}
	if req.Limit > 0 && req.Limit <= 100 {
		limit = req.Limit
	}

	resp, err := h.commentSvc.GetMoreReplies(ctx.Request.Context(), &commentv1.GetMoreRepliesRequest{
		Rid:   rid,
		MinId: minId,
		Limit: limit,
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
	}

	// 收集所有 uid 并批量查询用户名
	uidSet := make(map[int64]struct{})
	for _, r := range resp.GetReplies() {
		uidSet[r.Uid] = struct{}{}
	}
	userNameMap := make(map[int64]string)
	userAvatarMap := make(map[int64]string)
	for uid := range uidSet {
		userResp, er := h.userSvc.Profile(ctx.Request.Context(), &userv1.ProfileRequest{Id: uid})
		if er == nil && userResp.GetUser() != nil {
			userNameMap[uid] = userResp.GetUser().GetNickName()
			userAvatarMap[uid] = userResp.GetUser().GetAvatarUrl()
		}
	}

	replies := make([]Comment, 0, len(resp.GetReplies()))
	for _, r := range resp.GetReplies() {
		var pid int64 = 0
		var rootId int64 = 0
		if r.ParentComment != nil {
			pid = r.ParentComment.Id
		}
		if r.RootComment != nil {
			rootId = r.RootComment.Id
		}
		replies = append(replies, Comment{
			Id:            r.Id,
			Content:       r.Content,
			Uid:           r.Uid,
			UserName:      userNameMap[r.Uid],
			UserAvatarUrl: userAvatarMap[r.Uid],
			ParentId:      pid,
			RootId:        rootId,
			Ctime:         r.Ctime.AsTime().UnixMilli(),
		})
	}

	return ginx.Result{
		Code: 2,
		Msg:  "ok",
		Data: replies,
	}, nil
}

// ListLiked 查询用户点赞过的文章列表
func (h *ArticleHandler) ListLiked(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	var req ListLikedReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		return ginx.Result{Code: 4, Msg: "参数错误"}, err
	}
	biz := req.Biz
	if biz == "" {
		biz = h.biz
	}
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	resp, err := h.intrSvc.ListUserLiked(ctx.Request.Context(), &intrv1.ListUserLikedRequest{
		UserId: uc.UserId,
		Biz:    biz,
		Offset: req.Offset,
		Limit:  limit,
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
	}
	return ginx.Result{
		Code: 2,
		Msg:  "ok",
		Data: resp.GetBizIds(),
	}, nil
}

// ListCollected 查询用户收藏过的文章列表
func (h *ArticleHandler) ListCollected(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	var req ListCollectedReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		return ginx.Result{Code: 4, Msg: "参数错误"}, err
	}
	biz := req.Biz
	if biz == "" {
		biz = h.biz
	}
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	resp, err := h.intrSvc.ListUserCollected(ctx.Request.Context(), &intrv1.ListUserCollectedRequest{
		UserId: uc.UserId,
		Biz:    biz,
		Offset: req.Offset,
		Limit:  limit,
	})
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
	}
	return ginx.Result{
		Code: 2,
		Msg:  "ok",
		Data: resp.GetBizIds(),
	}, nil
}
