package ratelimit

import (
	"go.uber.org/atomic"
)

type CounterLimiter struct {
	cnt       atomic.Value
	threshold int32
}

//func (c *CounterLimiter) Limit(ctx context.Context, key string) (bool, error) {
//	cnt := c.cnt.Load().(int32)
//	if cnt > c.threshold {
//		return false, nil
//	}
//
//}
