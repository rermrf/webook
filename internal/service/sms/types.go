package sms

import "context"

type Service interface {
	// appId string, signature string
	Send(ctx context.Context, tpl string, args []string, numbers ...string) error
}

type NameArg struct {
	Val  string
	Name string
}
