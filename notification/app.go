package main

import (
	"webook/notification/scheduler"
	"webook/pkg/grpcx"
	"webook/pkg/saramax"
)

type App struct {
	server    *grpcx.Server
	consumers []saramax.Consumer
	scheduler *scheduler.CheckBackScheduler
}
