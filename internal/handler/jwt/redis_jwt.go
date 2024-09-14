package jwt

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	AccessTokenKey  = []byte("pBMH@cKP65sknQI%ijB2DzhFnvsfiyt*")
	RefreshTokenKey = []byte("IfL$*Xhqa*RBij5j@zF9$x*bcN8lMSZc")
)

type RedisJWTHandler struct {
	cmd redis.Cmdable
}

func NewRedisJWTHandler(cmd redis.Cmdable) Handler {
	return &RedisJWTHandler{
		cmd: cmd,
	}
}

func (r *RedisJWTHandler) ExtractToken(ctx *gin.Context) string {
	// 使用 JWT 进行登录校验
	tokenHeader := ctx.GetHeader("Authorization")
	segs := strings.Split(tokenHeader, " ")
	if len(segs) != 2 {
		// 没登录
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return ""
	}
	return segs[1]
}

func (r *RedisJWTHandler) SetJWTToken(ctx *gin.Context, userId int64, ssid string) error {
	claims := UserClaims{
		UserId:    userId,
		UserAgent: ctx.Request.UserAgent(),
		Ssid:      ssid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(AccessTokenKey)
	if err != nil {
		return err
	}
	log.Println(tokenStr)
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (r *RedisJWTHandler) ClearToken(ctx *gin.Context) error {
	// 退出登录将前端 header 设置为非法值
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")

	claims := ctx.MustGet("cliams").(*UserClaims)
	return r.cmd.Set(ctx, fmt.Sprintf("users:ssid:%s", claims.Ssid), "", time.Hour*24*7).Err()
}

func (r *RedisJWTHandler) SetLoginToken(ctx *gin.Context, userId int64) error {
	ssid := uuid.New().String()
	err := r.SetJWTToken(ctx, userId, ssid)
	if err != nil {
		return err
	}
	err = r.SetRefreshToken(ctx, userId, ssid)
	return err
}

func (r *RedisJWTHandler) SetRefreshToken(ctx *gin.Context, userId int64, ssid string) error {
	claims := RefreshClaims{
		UserId: userId,
		Ssid:   ssid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(RefreshTokenKey)
	if err != nil {
		return err
	}
	log.Println(tokenStr)
	ctx.Header("x-refresh-token", tokenStr)
	return nil
}

func (r *RedisJWTHandler) CheckSession(ctx *gin.Context, ssid string) error {
	_, err := r.cmd.Exists(ctx, fmt.Sprintf("users:ssid:%s", ssid)).Result()
	return err
}

func (r *RedisJWTHandler) setJWTToken(ctx *gin.Context, userId int64, ssid string) error {
	claims := UserClaims{
		UserId:    userId,
		UserAgent: ctx.Request.UserAgent(),
		Ssid:      ssid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(AccessTokenKey)
	if err != nil {
		return err
	}
	log.Println(tokenStr)
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

type RefreshClaims struct {
	UserId int64
	Ssid   string
	jwt.RegisteredClaims
}

type UserClaims struct {
	UserId    int64  `json:"userId"`
	Ssid      string `json:"ssid"`
	UserAgent string `json:"userAgent"`
	jwt.RegisteredClaims
}
