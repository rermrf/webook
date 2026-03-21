package main

import (
	"github.com/robfig/cron/v3"
	"webook/pkg/grpcx"
	"webook/pkg/saramax"
)

type App struct {
	server    *grpcx.Server
	consumers []saramax.Consumer
	cron      *cron.Cron
}
