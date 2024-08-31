package middleware

import (
	"encoding/gob"
	"log"
	"net/http"
	"strings"
	"time"
	"webook/internal/handler"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWT 登录校验
type LoginJWTMiddlewareBuilder struct {
	paths []string
}

func NewLoginJWTMiddlewareBuilder() *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{}
}

func (l *LoginJWTMiddlewareBuilder) IgnorePaths(path string) *LoginJWTMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {
	// 用 Go 的方式编码解码
	gob.Register(time.Now())
	return func(ctx *gin.Context) {
		// 不需要登录校验
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				ctx.Next()
				return
			}
		}
		// 使用 JWT 进行登录校验
		tokenHeader := ctx.GetHeader("Authorization")
		if tokenHeader == "" {
			// 没登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		segs := strings.Split(tokenHeader, " ")
		if len(segs) != 2 {
			// 没登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenStr := segs[1]
		claims := &handler.UserClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("pBMH@cKP65sknQI%ijB2DzhFnvsfiyt*"), nil
		})
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// claims.ExpiresAt.Time.Before(time.Now())
		// err 为 nil,token 不为 nil
		if token == nil || !token.Valid || claims.UserId == 0 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 校验 UserAgent
		if claims.UserAgent != ctx.Request.UserAgent() {
			// 严重的安全问题
			// 理论上要加监控
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 每十秒钟刷新一次
		now := time.Now()
		if claims.ExpiresAt.Sub(now) < 50*time.Second {
			claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute))
			tokenStr, err = token.SignedString([]byte("pBMH@cKP65sknQI%ijB2DzhFnvsfiyt*"))
			if err != nil {
				// 记录日志
				log.Println("jwt 续约失败", err)
			}
			ctx.Header("x-jwt-token", tokenStr)
		}

		ctx.Set("claims", claims)
		// ctx.Set("userId", claims.UserId)
	}
}
