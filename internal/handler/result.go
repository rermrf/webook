package handler

type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	data any    `json:"data"`
}
