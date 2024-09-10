package auth

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"webook/internal/service/sms"
)

type SMSService struct {
	svc sms.Service
	key string
}

// Send 发送， 其中 biz 必须是线下申请的一个代表业务方的token
func (s SMSService) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	// 权限校验
	// 如果这里能解析成功，说明就是对应的业务方
	// 没有 error 就说明，token 是我发的
	var tc Claims
	token, err := jwt.ParseWithClaims(biz, &tc, func(token *jwt.Token) (interface{}, error) {
		return s.key, nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return errors.New("token 不合法")
	}
	return s.svc.Send(ctx, tc.Tpl, args, numbers...)
}

type Claims struct {
	jwt.RegisteredClaims
	Tpl string `json:"tpl"`
}
