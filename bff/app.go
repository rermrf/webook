package main

import (
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"webook/pkg/saramax"
)

type App struct {
	Server    *gin.Engine
	Consumers []saramax.Consumer
	cron      *cron.Cron
}
