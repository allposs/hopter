package hopter

import (
	"github.com/allposs/hopter/monitor"
)

// metricMiddleware Metric插件
func (w *Web) metricMiddleware() {
	// get global Monitor object
	m := monitor.GetMonitor()
	// +optional set metric path, default /debug/metrics
	m.SetMetricPath("/metrics")
	// +optional set slow time, default 5s
	m.SetSlowTime(10)
	// +optional set request duration, default {0.1, 0.3, 1.2, 5, 10}
	// used to p95, p99
	m.SetDuration([]float64{0.1, 0.3, 1.2, 5, 10})
	// set middleware for gin
	m.Use(w.Engine)
	w.beanFactory.setBean(m)
}
