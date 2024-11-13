package wrr

import (
	"context"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"sync"
)

const name = "custom_wrr"

// balancer.Balancer 接口
// balancer.Picker 接口
// base.PickerBuilder 接口
func init() {
	// NewBalancerBuilder 是帮我们把一个 Picker Builder 转化为一个 balancer builder
	balancer.Register(base.NewBalancerBuilder(name, &PickerBuilder{}, base.Config{HealthCheck: true}))
}

//传统版本的基于权重的负载均衡算法

type PickerBuilder struct {
}

func (p *PickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	var conns []*conn
	for sc, scInfo := range info.ReadySCs {
		cc := &conn{
			cc: sc,
		}
		// metadata 是个被弃用的字段，所以有一些服务注册与发现就会把 metadata 里面的数据放到 Address 里面的 Attributes
		md, ok := scInfo.Address.Metadata.(map[string]any)
		if ok {
			weightVal := md["weight"]
			weight, _ := weightVal.(float64)
			cc.weight = int(weight)
			//cc.group = md["group"]
		}

		if cc.weight == 0 {
			cc.weight = 50
		}
		cc.currentWeight = cc.weight
		conns = append(conns, cc)

	}
	return &Picker{
		conns: conns,
	}
}

// Picker 这个才是真的执行负载均衡的地方
type Picker struct {
	conns []*conn
	mutex sync.Mutex
}

// Pick 在这里实现基于权重的负载均衡算法
func (p *Picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if len(p.conns) == 0 {
		// 没有候选节点
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	var total int
	var maxCC *conn
	// 计算当前权重
	for _, cc := range p.conns {
		if !cc.available {
			continue
		}
		// 性能最好就是在 cc 上用原子操作
		// 但是筛选结果不会严格符合 WRR 算法
		// 整体效果可以
		total += cc.weight
		cc.currentWeight += cc.weight
		if maxCC == nil || cc.currentWeight > maxCC.currentWeight {
			maxCC = cc
		}
	}
	// 更新
	maxCC.currentWeight -= total
	// maxCC 就是挑出来的
	return balancer.PickResult{
		SubConn: maxCC.cc,
		Done: func(info balancer.DoneInfo) {
			// 很多动态算法，根据通用结果来调整权重，就在这里
			err := info.Err
			if err == nil {
				// 可以考虑增加权重
				return
			}
			switch err {
			// 一般是主动取消，你没必要去调
			case context.Canceled:
				return
			case context.DeadlineExceeded:
			// 可以考虑降低权重
			case io.EOF, io.ErrUnexpectedEOF:
				// 基本可以认为这个节点已经崩了
			default:
				st, ok := status.FromError(err)
				if ok {
					code := st.Code()
					switch code {
					case codes.Unavailable:
						// 这里可能表达的是熔断
						// 就要考虑挪走该节点，这个节点已经不可用了
						// 注意并发问题，可以使用原子操作
						maxCC.available = false
						go func() {
							// 额外开一个 goroutine 去探活
							// 借助 health check
							// for 循环
							if p.healthCheck(maxCC) {
								// 放回来
								maxCC.available = true
								// 最好加点流量控制的措施
								// 要求下一次选中 maxCC 的时候，掷骰子，生成一随机数，如果大于特定的阈值，就将这个请求发过去

							}
						}()
					case codes.ResourceExhausted:
						// 这里可能表达的是限流
						// 可以挪走
						// 也可以留着，留着的话，就要降低权重，最好是 currentWeight 和 weight 都调低
						// 减少它被选中的概率

						// 加一个错误吗表达降级
					}
				}
			}
		},
	}, nil
}

func (p *Picker) healthCheck(cc *conn) bool {
	// 调用 grpc 内置的 health check 接口
	return true
}

// conn 代表节点
type conn struct {
	// 初始权重
	weight int
	// 有效权重
	efficientWeight int
	currentWeight   int

	// 真正的，grpc 里面的代表一个节点的表达
	cc balancer.SubConn

	available bool

	// 假如有 vip 或者 非vip
	group string
}
