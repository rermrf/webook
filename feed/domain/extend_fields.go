package domain

import (
	"errors"
	"fmt"
)

type ExtendFields map[string]string

var errKeyNotFound = errors.New("没有找到对应的 key")

func (f ExtendFields) Get(key string) Result {
	val, ok := f[key]
	if !ok {
		return Result{
			Err: fmt.Errorf("%w,key %s", errKeyNotFound, key),
		}
	}
	return Result{
		Val: val,
	}
}

type Result struct {
	Val any
	Err error
}
