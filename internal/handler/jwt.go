package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"time"
)

type JWTHandler struct {
}

func (h JWTHandler) setJWTToken(ctx *gin.Context, userId int64) error {
	claims := UserClaims{
		UserId:    userId,
		UserAgent: ctx.Request.UserAgent(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte("pBMH@cKP65sknQI%ijB2DzhFnvsfiyt*"))
	if err != nil {
		return err
	}
	log.Println(tokenStr)
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

type UserClaims struct {
	UserId    int64  `json:"userId"`
	UserAgent string `json:"userAgent"`
	jwt.RegisteredClaims
}
