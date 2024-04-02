package hopter

import (
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
func (w *Web) metric() {
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
	m.Use(w.engine)
	w.beanFactory.set(m)
}

// Attach 中间件加入
func (w *Web) Attach(m ...Middleware) *Web {
	for _, v := range m {
		w.engine.Use(func(ctx *gin.Context) {
			w.beanFactory.Inject(v.OnInject())
			err := v.Handler(&Context{ctx})
			if err != nil {
				ctx.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
			} else {
				ctx.Next()
			}
		})
	}
	return w
}
