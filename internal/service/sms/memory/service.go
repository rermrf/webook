package memory

import (
	"context"
	"log"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s Service) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	log.Println(args)
	return nil
}
