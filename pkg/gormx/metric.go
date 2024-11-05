package gormx

import (
	promsdk "github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"
	"time"
)

// PrometheusSummaryMetricPlugin 一个 gorm 的 prometheus 的插件，对 sql 的执行时间进行统计
type PrometheusSummaryMetricPlugin struct {
	nameSpace string
	subsystem string
	name      string
	help      string
	db        string
	vector    *promsdk.SummaryVec
}

func (c *PrometheusSummaryMetricPlugin) Name() string {
	return "promethus-metrics"
}

func (c *PrometheusSummaryMetricPlugin) Initialize(db *gorm.DB) error {
	c.registerAll(db)
	return nil
}

func NewPrometheusSummaryMetricPlugin(nameSpace, subsystem, name, help, db string) *PrometheusSummaryMetricPlugin {
	vector := promsdk.NewSummaryVec(promsdk.SummaryOpts{
		// 在这边要考虑设置各种 Namespace
		Namespace: nameSpace,
		Subsystem: subsystem,
		Name:      name,
		Help:      help,
		ConstLabels: map[string]string{
			"db": db,
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	},
		// 如果是 JOIN 查询，table 就是 JOIN 在一起的
		// 或者 table 就是主表，A JOIN B，记录的是 A
		[]string{"type", "table"})

	pcb := &PrometheusSummaryMetricPlugin{
		vector: vector,
	}
	promsdk.MustRegister(vector)
	return pcb
}

func (c *PrometheusSummaryMetricPlugin) registerAll(db *gorm.DB) {
	// 作用于 INSERT 语句
	err := db.Callback().Create().Before("*").Register("prometheus_create_before", c.before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Create().After("*").Register("prometheus_create_after", c.after("create"))
	if err != nil {
		panic(err)
	}

	err = db.Callback().Update().Before("*").Register("prometheus_update_before", c.before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Update().After("*").Register("prometheus_update_after", c.after("update"))
	if err != nil {
		panic(err)
	}

	err = db.Callback().Delete().Before("*").Register("prometheus_delete_before", c.before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Delete().After("*").Register("prometheus_delete_after", c.after("delete"))
	if err != nil {
		panic(err)
	}

	err = db.Callback().Query().Before("*").Register("prometheus_query_before", c.before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Query().After("*").Register("prometheus_query_after", c.after("query"))
	if err != nil {
		panic(err)
	}

	err = db.Callback().Row().Before("*").Register("prometheus_row_before", c.before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Row().After("*").Register("prometheus_row_after", c.after("row"))
	if err != nil {
		panic(err)
	}

	err = db.Callback().Raw().Before("*").Register("prometheus_row_before", c.before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Raw().After("*").Register("prometheus_row_after", c.after("raw"))
	if err != nil {
		panic(err)
	}
}

func (c *PrometheusSummaryMetricPlugin) before() func(db *gorm.DB) {
	return func(db *gorm.DB) {
		startTime := time.Now()
		db.Set("start_time", startTime)
	}
}

func (c *PrometheusSummaryMetricPlugin) after(typ string) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		val, _ := db.Get("strat_time")
		startTime, ok := val.(time.Time)
		if !ok {
			// 啥都干不了
			return
		}
		// 准备上报 prometheus

		table := db.Statement.Table
		if table == "" {
			table = "unknown"
		}
		c.vector.WithLabelValues(typ, table).Observe(float64(time.Since(startTime).Milliseconds()))
	}
}
