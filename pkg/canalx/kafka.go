package canalx

// Message 可以根据需要把其他字段加入进来
// T 直接对应到表结构
type Message[T any] struct {
	Data     []T    `json:"data"`
	Database string `json:"database"`
	Table    string `json:"table"`
	Type     string `json:"type"`
}
