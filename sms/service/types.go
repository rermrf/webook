package service

import "context"

//go:generate mockgen -source=./types.go -package=smsmocks -destination=./mocks/sms_mock.go
type Service interface {
	// Send biz 很含糊的业务
	Send(ctx context.Context, biz string, args []string, numbers ...string) error
}

type NameArg struct {
	Val  string
	Name string
}
