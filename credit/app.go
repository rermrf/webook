package main

import (
	"webook/pkg/ginx"
	"webook/pkg/grpcx"
	"webook/pkg/saramax"
)

type App struct {
	server     *grpcx.Server
	httpServer *ginx.Server
	consumers  []saramax.Consumer
}
