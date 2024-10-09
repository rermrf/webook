package handler

import (
	"webook/internal/pkg/logger"
	"webook/internal/service"
)

// vo: view object, 对标前端的

type ArticleVO struct {
	Id       int64  `json:"id"`
	Title    string `json:"title"`
	Abstract string `json:"abstract"`
	Content  string `json:"content"`
	// Author 要从用户来
	AuthorId   int64
	AuthorName string
	// 状态可以是前端来处理，也可以是后端来处理
	// 0 -> unknown -> 未知状态
	// 1 -> unpublish -> 未发表状态
	// 2 -> publish -> 已发表
	// 3 -> private -> 私密
	Status uint8
	Ctime  string
	Utime  string
}

type ArticleReq struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type ListReq struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type WithdrawReq struct {
	Id int64
}

type ArticleHandler struct {
	svc service.ArticleService
	l   logger.LoggerV1
}
