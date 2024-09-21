package gin_pulgin

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"webook/internal/handler"
)

func WarpReq[T any](fn func(ctx *gin.Context, req T) (handler.Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req T
		if err := ctx.Bind(&req) != nil; err {
			return
		}
		res, err := fn(ctx, req)
		if err != nil {

		}
		ctx.JSON(http.StatusOK, res)
	}
}