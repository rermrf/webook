package api

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrAppNotFound         = errors.New("app not found")
	ErrAppDisabled         = errors.New("app is disabled")
	ErrInvalidSign         = errors.New("invalid signature")
	ErrTimestampExpired    = errors.New("timestamp expired")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidIPAddress    = errors.New("invalid IP address")
)

// SignParams 签名参数
type SignParams struct {
	AppId     string
	Timestamp int64
	Nonce     string
	Body      string // 请求体
}

// GenerateTradeNo 生成交易号
func GenerateTradeNo(appId string) string {
	return fmt.Sprintf("credit-%s-%d", appId, time.Now().UnixNano())
}

// Result 通用响应结构
type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}
