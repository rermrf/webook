package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type JobDao interface {
	Preempt(ctx context.Context) (Job, error)
	Release(ctx context.Context, id int64) error
	UpdateUtime(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, next time.Time) error
	Stop(ctx context.Context, id int64) error
}

type GORMJobDAO struct {
	db *gorm.DB
}

func NewGORMJobDAO(db *gorm.DB) JobDao {
	return &GORMJobDAO{db: db}
}

func (dao *GORMJobDAO) Preempt(ctx context.Context) (Job, error) {
	// 高并发情况下，大部分都是陪跑
	// 100 个 goroutine
	// 所有 goroutine 执行的循环次数加在一起是
	// 100！
	// 特定一个 goroutine，最差情况下，要循环一百次
	db := dao.db.WithContext(ctx).Model(&Job{})
	for {
		now := time.Now().UnixMilli()
		var j Job
		// 分布式任务调度系统
		// 1. 一次拉一批，一次性取出 100 条来，然后随机从某一条开始，向后开始抢占
		// 2. 随机偏移量，0-100 生成一个随机偏移量。兜底：第一轮没查到，偏移量回归到 0
		// 3. id 取余分配， status = ? AND next_time <= ? AND id%10 = ? 兜底：不加余数条件，取 next_time 最老的

		err := db.Where("status = ? AND next_time <= ?", jobStatusWaiting, now).First(&j).Error
		// 找到了可以被抢占的任务
		// 找到了之后，就要抢占了
		if err != nil {
			return Job{}, err
		}
		// 两个 goroutine 都拿到 id = 1 的数据
		// 乐观锁，CAS 操作，compare and swap
		// 有一个很常见的面试刷亮点：使用乐观锁取代 FOR UPDATE
		// 面试套路（性能优化）：曾经用了 FOR UPDATE => 性能差，还会有死锁 => 优化成了乐观锁
		res := db.Where("id=? AND version = ?", j.Id, j.Version).Updates(map[string]interface{}{
			"status":  jobStatusRunning,
			"utime":   now,
			"version": j.Version + 1,
		})
		if res.Error != nil {
			return Job{}, res.Error
		}
		if res.RowsAffected == 0 {
			// 抢占失败，只能继续下一轮
			//return Job{}, errors.New("没抢到")
			continue

		}
		return j, nil
	}
}

func (dao *GORMJobDAO) Release(ctx context.Context, id int64) error {
	// 这里有一个问题，要不要检测 status 或者 version？
	// WHERE version = ?
	// 防止释放掉别人的任务
	return dao.db.WithContext(ctx).Model(&Job{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status": jobStatusWaiting,
		"utime":  time.Now().UnixMilli(),
	}).Error
}

func (dao *GORMJobDAO) UpdateUtime(ctx context.Context, id int64) error {
	return dao.db.WithContext(ctx).Model(&Job{}).Where("id = ?", id).Updates(map[string]interface{}{
		"utime": time.Now().UnixMilli(),
	}).Error
}

func (dao *GORMJobDAO) UpdateNextTime(ctx context.Context, id int64, next time.Time) error {
	return dao.db.WithContext(ctx).Model(&Job{}).Where("id = ?", id).Updates(map[string]interface{}{
		"next_time": next.UnixMilli(),
		"utime":     time.Now().UnixMilli(),
	}).Error
}

func (dao *GORMJobDAO) Stop(ctx context.Context, id int64) error {
	return dao.db.WithContext(ctx).Model(&Job{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status": jobStatusPaused,
		"utime":  time.Now().UnixMilli(),
	}).Error
}

type Job struct {
	Id       int64 `gorm:"primary_key,autoIncrement"`
	Cfg      string
	Name     string `gorm:"unique"`
	Executor string

	// 第一个问题：哪些任务可以抢？哪些任务已经被人占着？哪些任务永远不会被运行?
	// 用状态来标记
	Status int

	// 另外一个问题，怎么知道已经到时间了
	// NextTime 下一次被调度的时间
	// next_time <= now 这样一个查询条件
	// and status = 0
	// 更加好的应该是好 status 和 next_time 的联合索引
	NextTime int64 `gorm:"index"`
	// cron 表达式
	Cron string

	Version int

	Ctime int64
	Utime int64
}

const (
	jobStatusWaiting = iota
	// 已经被抢占
	jobStatusRunning
	// 暂停调度
	jobStatusPaused
)
