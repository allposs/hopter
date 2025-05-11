package hopter

import (
	"runtime/debug"

	"github.com/allposs/hopter/metric"
	"github.com/gin-gonic/gin"
)

// Middleware 中间件接口
type Middleware interface {
	// Handler 处理方法
	Handler(ctx *Context) error
	//OnInject 用于对象注入
	OnInject() any
}

// metric Metric插件
func (e *Engine) metric() {
	// get global Monitor object
	m := metric.GetMonitor()
	// +optional set metric path, default /debug/metrics
	m.SetMetricPath("/metrics")
	// +optional set slow time, default 5s
	m.SetSlowTime(10)
	// +optional set request duration, default {0.1, 0.3, 1.2, 5, 10}
	// used to p95, p99
	m.SetDuration([]float64{0.1, 0.3, 1.2, 5, 10})
	// set middleware for gin
	m.Use(e.engine)
	e.beanFactory.set(m)
}

// Attach 中间件加入
func (e *Engine) Attach(m ...Middleware) *Engine {
	for _, v := range m {
		e.engine.Use(func(ctx *gin.Context) {
			e.beanFactory.Inject(v.OnInject())
			err := v.Handler(&Context{ctx})
			if err != nil {
				ctx.AbortWithStatusJSON(400, gin.H{"web服务异常:%v,请联系管理人员": err})
				debug.PrintStack()
			} else {
				ctx.Next()
			}
		})
	}
	return e
}
