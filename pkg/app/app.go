package app

import (
	"webook/pkg/grpcx"
	"webook/pkg/saramax"
)

type App struct {
	Server    *grpcx.Server
	Consumers []saramax.Consumer
}
