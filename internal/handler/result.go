package handler

type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"Data"`
}

//// 重构小技巧，直接使用别名
//type Result = gin_pulgin.Result
