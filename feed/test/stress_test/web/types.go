package web

import "webook/pkg/ginx"

type FindFeedEventReq struct {
	Uid       int64 `json:"uid"`
	Limit     int64 `json:"limit"`
	Timestamp int64 `json:"timestamp"`
}

type CreateFeedEventReq struct {
	Typ string `json:"typ"`
	Ext string `json:"ext"`
}

type Result = ginx.Result
