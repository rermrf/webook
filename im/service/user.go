package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"webook/im/domain"
)

const defaultBase = "http://localhost:10002/user/user_register"

type UserSercie interface {
	Sync(ctx context.Context, user domain.User) error
}

type RestUserService struct {
	// 部署 IM 时候配置的 IM Secret，默认是 openIM123
	secret string
	base   string
	client *http.Client
}

func NewRestUserService(secret string, base string) UserSercie {
	if base == "" {
		base = defaultBase
	}
	// 假如我有 TLS 之类的认证，我在这里可以灵活替换具体的 client
	return &RestUserService{secret: secret, base: base, client: http.DefaultClient}
}

func (s *RestUserService) Sync(ctx context.Context, user domain.User) error {
	// 调用 openIM 的接口
	spanCtx := trace.SpanContextFromContext(ctx)
	var traceId string
	if spanCtx.HasSpanID() {
		traceId = spanCtx.SpanID().String()
	} else {
		// 随便生成一个，但是这样链路拼接不起来了
		traceId = uuid.New().String()
	}
	body := syncUserRequest{
		Secret: s.secret,
		Users:  []User{{UserId: user.UserId, Nickname: user.Nickname, FaceURL: user.FaceURL}},
	}
	bodyVal, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.base+"/user/user_register", bytes.NewBuffer(bodyVal))
	if err != nil {
		return err
	}
	req.Header.Set("operationID", traceId)
	res, err := s.client.Do(req)
	if err != nil {
		return err
	}
	var resp response
	_ = json.NewDecoder(res.Body).Decode(&resp)
	if resp.ErrCode != 0 {
		return fmt.Errorf("数据同步失败 %d, %s, %s", resp.ErrCode, resp.ErrMsg, resp.ErrDlt)
	}
	return nil
}

type syncUserRequest struct {
	Secret string `json:"secret"`
	Users  []User `json:"users"`
}

type response struct {
	// 错误码，0 表示成功
	ErrCode int `json:"errCode"`
	// 错误简要信息，为空
	ErrMsg string `json:"errMsg"`
	// 错误详细信息，为空
	ErrDlt string `json:"errDlt"`
}

type User struct {
	UserId   string `json:"userId"`
	Nickname string `json:"nickname"`
	FaceURL  string `json:"faceUrl"`
}
