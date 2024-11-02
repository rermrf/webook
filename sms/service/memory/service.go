package memory

import (
	"context"
	"log"
	"webook/pkg/logger"
)

type Service struct {
	l logger.LoggerV1
}

func NewService(l logger.LoggerV1) *Service {
	return &Service{
		l: l,
	}
}

func (s Service) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	log.Println(args)
	return nil
}
