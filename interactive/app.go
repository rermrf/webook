package main

import (
	"webook/pkg/ginx"
	"webook/pkg/grpcx"
	"webook/pkg/saramax"
)

type App struct {
	// 在这里，所有需要 main 函数控制启动、关闭的都会在这里有一个
	// 核心就是为了控制生命周期
	server    *grpcx.Server
	consumers []saramax.Consumer
	webAdmin  *ginx.Server
}
