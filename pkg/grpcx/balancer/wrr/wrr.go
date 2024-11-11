package wrr

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
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
		},
	}, nil
}

// conn 代表节点
type conn struct {
	// 权重
	weight        int
	currentWeight int

	// 真正的，grpc 里面的代表一个节点的表达
	cc balancer.SubConn
}
