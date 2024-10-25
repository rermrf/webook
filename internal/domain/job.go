package domain

import (
	"github.com/robfig/cron/v3"
	"time"
)

type Job struct {
	Id int64
	// 比如说 ranking
	Name string

	Cron     string
	Exectuor string
	// 通用的任务的抽象，我们也不知道任务具体细节，所以搞一个 Cfg
	// 具体任务设置具体的值
	Cfg string

	CancelFunc func() error
}

var parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

func (j Job) NextTime() time.Time {
	// 下一次执行时间要根据cron表达式来算
	s, _ := parser.Parse(j.Cron)
	return s.Next(time.Now())
}
