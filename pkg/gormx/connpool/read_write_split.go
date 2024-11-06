package connpool

import (
	"context"
	"database/sql"
	"gorm.io/gorm"
)

// ReadWriteSplit 读写分离、主从模式 装饰器模式，实现 ConnPool、TxBeginner 接口
type ReadWriteSplit struct {
	master gorm.ConnPool
	slaves []gorm.ConnPool
}

// BeginTx 开事物只可能在 master 上开事物
func (r *ReadWriteSplit) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return r.master.(gorm.TxBeginner).BeginTx(ctx, opts)
}

func (r *ReadWriteSplit) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	// 可以默认返回 master，也可以默认返回 slave
	return r.master.PrepareContext(ctx, query)
}

func (r *ReadWriteSplit) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	// 写操作都是走 master
	return r.master.ExecContext(ctx, query, args...)
}

func (r *ReadWriteSplit) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	// slaves 要考虑负载均衡，轮训
	// 这里可以玩花活，轮训、加权轮训、平滑的加权轮训、随机、加权随机
	// 动态判定 slaves 健康情况的负载均衡策略（永远挑最快返回响应的那个 slave，或者暂时性禁用超时的slaves）
	panic("")
}

func (r *ReadWriteSplit) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	//TODO implement me
	panic("implement me")
}
