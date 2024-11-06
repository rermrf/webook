package dao

import (
	"context"
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

type DoubleWrite struct {
	src     InteractiveDao
	dst     InteractiveDao
	pattern atomic.String
}

func NewDoubleWriteV1(src *gorm.DB, dst *gorm.DB) *DoubleWrite {
	pattern := atomic.String{}
	pattern.Store(patternSrcOnly)
	return &DoubleWrite{src: NewGORMInteractiveDao(src), dst: NewGORMInteractiveDao(dst), pattern: pattern}
}

func NewDoubleWrite(src InteractiveDao, dst InteractiveDao) *DoubleWrite {
	return &DoubleWrite{src: src, dst: dst}
}

func (d *DoubleWrite) UpdatePattern(pattern string) {
	d.pattern.Store(pattern)
}

func (d *DoubleWrite) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	switch d.pattern.Load() {
	case patternSrcOnly:
		return d.src.IncrReadCnt(ctx, biz, bizId)
	case patternSrcFirst:
		err := d.src.IncrReadCnt(ctx, biz, bizId)
		if err != nil {
			// 要不要继续写 dst
			// 万一，我的err是超时错误呢？
			return err
		}
		// 万一，SRC成功了，但是 DST 失败了呢？
		// 只能等校验与修复
		err = d.dst.IncrReadCnt(ctx, biz, bizId)
		if err != nil {
			// 记录日志
			// dst 写失败，不被认为是失败
		}
		return nil
	case patternDstFirst:
		err := d.dst.IncrReadCnt(ctx, biz, bizId)
		if err != nil {
			return err
		}
		err = d.src.IncrReadCnt(ctx, biz, bizId)
		if err != nil {
			// 记录日志
		}
		return nil
	case patternDstOnly:
		return d.dst.IncrReadCnt(ctx, biz, bizId)
	default:
		return errors.New("未知的双写模式")
	}
}

func (d *DoubleWrite) InsertLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWrite) DeleteLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWrite) InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWrite) DeleteCollectionBiz(ctx context.Context, biz string, bizId int64, uid int64) error {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWrite) GetCollectInfo(ctx context.Context, biz string, bizId int64, uid int64) (UserCollectionBiz, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWrite) GetLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) (UserLikeBiz, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWrite) Get(ctx context.Context, biz string, bizId int64) (Interactive, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWrite) BatchIncrReadCnt(ctx context.Context, bizs []string, ids []int64) error {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWrite) GetByIds(ctx context.Context, biz string, ids []int64) ([]Interactive, error) {
	//TODO implement me
	panic("implement me")
}
