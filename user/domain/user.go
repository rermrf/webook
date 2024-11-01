package domain

import (
	"time"
)

// User 领域对象，是 DDD 中的 entity
type User struct {
	Id       int64
	Email    string
	Nickname string
	Phone    string
	Password string
	// 不要组合，万一你将来可能还有 DingDingInfo，里面有同名字段
	WechatInfo WechatInfo
	AboutMe    string
	Ctime      time.Time
	Birthday   time.Time
}
