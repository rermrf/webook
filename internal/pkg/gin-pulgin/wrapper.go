package gin_pulgin

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"webook/internal/domain"
	"webook/internal/pkg/logger"
)

// L 使用包变量
var L logger.LoggerV1

func WrapBodyAndToken[T any, C jwt.Claims](fn func(ctx *gin.Context, req T, uc C) (domain.Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req T
		if err := ctx.ShouldBind(&req); err != nil {
			return
		}

		val, ok := ctx.Get("claims")
		if ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c, ok := val.(C)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 下半段业务逻辑
		// 业务逻辑也可能使用 ctx
		res, err := fn(ctx, req, c)
		if err != nil {
			// 开始处理 error，记录日志
			L.Error("处理业务逻辑出现错误",
				logger.String("path", ctx.Request.URL.Path),
				logger.String("route", ctx.FullPath()),
				logger.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapBodyV1[T any](fn func(ctx *gin.Context, req T) (domain.Result, error)) gin.HandlerFunc {
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
			L.Error("处理业务逻辑出现错误",
				logger.String("path", ctx.Request.URL.Path),
				logger.String("route", ctx.FullPath()),
				logger.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapBody[T any](l logger.LoggerV1, fn func(ctx *gin.Context, req T) (domain.Result, error)) gin.HandlerFunc {
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
