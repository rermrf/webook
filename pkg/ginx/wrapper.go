package ginx

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"strconv"
	ijwt "webook/internal/handler/jwt"
	"webook/pkg/logger"
)

// L 使用包变量
//var L logger.LoggerV1

var vector *prometheus.CounterVec

func InitCounter(opt prometheus.CounterOpts) {
	// 可以考虑使用 code，method，命中路由，HTTP 状态码
	vector = prometheus.NewCounterVec(opt, []string{"code"})
	prometheus.MustRegister(vector)
}

func WrapBodyAndToken[T any, C ijwt.UserClaims](l logger.LoggerV1, fn func(ctx *gin.Context, req T, uc C) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req T
		if err := ctx.ShouldBind(&req); err != nil {
			return
		}

		val, ok := ctx.Get("claims")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c, ok := val.(*C)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 下半段业务逻辑
		// 业务逻辑也可能使用 ctx
		res, err := fn(ctx, req, *c)
		if err != nil {
			// 开始处理 error，记录日志
			l.Error("处理业务逻辑出现错误",
				logger.String("path", ctx.Request.URL.Path),
				logger.String("route", ctx.FullPath()),
				logger.Error(err))
		}
		vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapBody[T any](l logger.LoggerV1, fn func(ctx *gin.Context, req T) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req T
		if err := ctx.ShouldBind(&req); err != nil {
			return
		}
		// 下半段业务逻辑
		// 业务逻辑也可能使用 ctx
		res, err := fn(ctx, req)
		if err != nil {
			// 开始处理 error，记录日志
			l.Error("处理业务逻辑出现错误",
				logger.String("path", ctx.Request.URL.Path),
				logger.String("route", ctx.FullPath()),
				logger.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapClaims[C ijwt.UserClaims](l logger.LoggerV1, fn func(ctx *gin.Context, uc C) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		val, ok := ctx.Get("claims")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		uc, ok := val.(*C)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		res, err := fn(ctx, *uc)
		if err != nil {
			l.Error("执行业务逻辑失败", logger.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
	}
}

//func WrapBodyV2[T any](fn func(ctx *gin.Context, req T) (handler.Result, error)) gin.HandlerFunc {
//	return func(ctx *gin.Context) {
//		var req T
//		if err := ctx.ShouldBind(&req); err != nil {
//			return
//		}
//		// 下半段业务逻辑
//		// 业务逻辑也可能使用 ctx
//		res, err := fn(ctx, req)
//		if err != nil {
//			// 开始处理 error，记录日志
//			L.Error("处理业务逻辑出现错误",
//				logger.String("path", ctx.Request.URL.Path),
//				logger.String("route", ctx.FullPath()),
//				logger.Error(err))
//		}
//		ctx.JSON(http.StatusOK, res)
//	}
//}
//
//func WrapBodyV1[T any](l logger.LoggerV1, fn func(ctx *gin.Context, req T) (handler.Result, error)) gin.HandlerFunc {
//	return func(ctx *gin.Context) {
//		var req T
//		if err := ctx.ShouldBind(&req); err != nil {
//			return
//		}
//		// 下半段业务逻辑
//		// 业务逻辑也可能使用 ctx
//		res, err := fn(ctx, req)
//		if err != nil {
//			// 开始处理 error，记录日志
//			l.Error("处理业务逻辑出现错误",
//				logger.String("path", ctx.Request.URL.Path),
//				logger.String("route", ctx.FullPath()),
//				logger.Error(err))
//		}
//		ctx.JSON(http.StatusOK, res)
//	}
//}
