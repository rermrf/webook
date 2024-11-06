package connpool

import (
	"context"
	"database/sql"
	"errors"
	"go.uber.org/atomic"
	"gorm.io/gorm"
)

const (
	patternSrcOnly  = "SRC_ONLY"
	patternSrcFirst = "SRC_FIRST"
	patternDstOnly  = "DST_ONLY"
	patternDstFirst = "DST_FIRST"
)

var (
	errUnknownPattern = errors.New("未知的双写模式")
)

type DoubleWritePool struct {
	src     gorm.ConnPool
	dst     gorm.ConnPool
	pattern atomic.String
}

func NewDoubleWritePool(src gorm.ConnPool, dst gorm.ConnPool) *DoubleWritePool {
	return &DoubleWritePool{src: src, dst: dst}
}

func (d *DoubleWritePool) UpdatePattern(pattern string) {
	d.pattern.Store(pattern)
}

func (d *DoubleWritePool) BeginTx(ctx context.Context, opts *sql.TxOptions) (gorm.ConnPool, error) {
	pattern := d.pattern.Load()
	switch pattern {
	case patternSrcOnly:
		tx, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		return &DoubleWritePoolTx{
			src:     tx,
			pattern: pattern,
		}, err
	case patternSrcFirst:
		srcTx, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		dstTx, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			// 记录日志，然后不做处理
			// 也可以考虑回滚
			//return nil,srcTx.Rollback()
		}
		return &DoubleWritePoolTx{
			src:     srcTx,
			dst:     dstTx,
			pattern: pattern,
		}, nil
	case patternDstOnly:
		tx, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		return &DoubleWritePoolTx{
			dst:     tx,
			pattern: pattern,
		}, err
	case patternDstFirst:
		dstTx, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		srcTx, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			// 记录日志，然后不做处理
		}
		return &DoubleWritePoolTx{
			src:     srcTx,
			dst:     dstTx,
			pattern: pattern,
		}, nil
	default:
		return nil, errUnknownPattern
	}
}

// PrepareContext Prepare 的语句会进来这里（指的是数据库的预编译语句）
func (d *DoubleWritePool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	// sql.Stmt 是一个结构体，你没有办法说返回一个代表双写的 Stmt
	//panic("implement me")
	return nil, errors.New("双写模式下不支持")
}

func (d *DoubleWritePool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	// 增删改，或者说，任何非查询语句，都会进来这里
	switch d.pattern.Load() {
	case patternSrcOnly:
		return d.src.ExecContext(ctx, query, args...)
	case patternSrcFirst:
		res, err := d.src.ExecContext(ctx, query, args...)
		if err != nil {
			return res, err
		}
		_, err = d.dst.ExecContext(ctx, query, args...)
		if err != nil {
			// 记录日志
		}
		return res, nil
	case patternDstFirst:
		res, err := d.dst.ExecContext(ctx, query, args...)
		if err != nil {
			return res, err
		}
		_, err = d.src.ExecContext(ctx, query, args...)
		if err != nil {
			// 记录日志
		}
		return res, nil
	case patternDstOnly:
		return d.dst.ExecContext(ctx, query, args...)
	default:
		panic("未知的双写模式")
		//return nil, errors.New("未知的双写模式")
	}
}

func (d *DoubleWritePool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	switch d.pattern.Load() {
	case patternSrcOnly, patternSrcFirst:
		return d.src.QueryContext(ctx, query, args...)
	case patternDstOnly, patternDstFirst:
		return d.dst.QueryContext(ctx, query, args...)
	default:
		panic("未知的双写模式")
		//return nil, errors.New("未知的双写模式")
	}
}

func (d *DoubleWritePool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	switch d.pattern.Load() {
	case patternSrcOnly, patternSrcFirst:
		return d.src.QueryRowContext(ctx, query, args...)
	case patternDstOnly, patternDstFirst:
		return d.dst.QueryRowContext(ctx, query, args...)
	default:
		// 构造不了 *sql.Row，只能 panic
		panic("未知的双写模式")
	}
}

type DoubleWritePoolTx struct {
	src *sql.Tx
	dst *sql.Tx
	// 事物是不能并行的，所以可以不用原子操作
	pattern string
}

func (d *DoubleWritePoolTx) Commit() error {
	switch d.pattern {
	case patternSrcOnly:
		return d.src.Commit()
	case patternSrcFirst:
		// 源表提交失败
		err := d.src.Commit()
		if err != nil {
			return err
		}
		if d.dst != nil {
			err = d.dst.Commit()
			if err != nil {
				// 记日志
			}
		}
		return nil
	case patternDstOnly:
		return d.dst.Commit()
	case patternDstFirst:
		// 目标表为空
		err := d.dst.Commit()
		if err != nil {
			return err
		}
		if d.src != nil {
			err = d.src.Commit()
			if err != nil {
				// 记日志
			}
		}
		return nil
	default:
		return errUnknownPattern
	}
}

func (d *DoubleWritePoolTx) Rollback() error {
	switch d.pattern {
	case patternSrcOnly:
		return d.src.Rollback()
	case patternSrcFirst:
		// 源表提交失败
		err := d.src.Rollback()
		if err != nil {
			// 可以尝试 rollback
			return err
		}
		if d.dst != nil {
			err = d.dst.Rollback()
			if err != nil {
				// 记日志
			}
		}
		return nil
	case patternDstOnly:
		return d.dst.Rollback()
	case patternDstFirst:
		// 目标表为空
		err := d.dst.Rollback()
		if err != nil {
			return err
		}
		if d.src != nil {
			err = d.src.Rollback()
			if err != nil {
				// 记日志
			}
		}
		return nil
	default:
		return errUnknownPattern
	}
}

func (d *DoubleWritePoolTx) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWritePoolTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	switch d.pattern {
	case patternSrcOnly:
		return d.src.ExecContext(ctx, query, args...)
	case patternSrcFirst:
		res, err := d.src.ExecContext(ctx, query, args...)
		if err != nil {
			return res, err
		}
		if d.dst == nil {
			return res, err
		}
		_, err = d.dst.ExecContext(ctx, query, args...)
		if err != nil {
			// 记录日志
		}
		return res, nil
	case patternDstFirst:
		res, err := d.dst.ExecContext(ctx, query, args...)
		if err != nil {
			return res, err
		}
		if d.src == nil {
			return res, err
		}
		_, err = d.src.ExecContext(ctx, query, args...)
		if err != nil {
			// 记录日志
		}
		return res, nil
	case patternDstOnly:
		return d.dst.ExecContext(ctx, query, args...)
	default:
		panic("未知的双写模式")
	}
}

func (d *DoubleWritePoolTx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	switch d.pattern {
	case patternSrcOnly, patternSrcFirst:
		return d.src.QueryContext(ctx, query, args...)
	case patternDstOnly, patternDstFirst:
		return d.dst.QueryContext(ctx, query, args...)
	default:
		panic("未知的双写模式")
		//return nil, errors.New("未知的双写模式")
	}
}

func (d *DoubleWritePoolTx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	switch d.pattern {
	case patternSrcOnly, patternSrcFirst:
		return d.src.QueryRowContext(ctx, query, args...)
	case patternDstOnly, patternDstFirst:
		return d.dst.QueryRowContext(ctx, query, args...)
	default:
		// 构造不了 *sql.Row，只能 panic
		panic("未知的双写模式")
	}
}
