package prometheus

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/zhang-yong-feng/webz"
)

type MiddlewareBuilder struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
}

func (m MiddlewareBuilder) Build() webz.HandleFunc {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:      m.Name, //让用户传入
		Subsystem: m.Subsystem,
		Namespace: m.Namespace,
		Help:      m.Help,
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.90:  0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}, []string{"pattern", "method", "status"})
	prometheus.MustRegister(vector)
	return func(ctx *webz.Context) {
		startime := time.Now()
		defer func() {
			duration := time.Since(startime).Microseconds()
			pattern := ctx.MathedRoute //路由名字
			if pattern == "" {
				pattern = "unknown"
			}
			vector.WithLabelValues(pattern, ctx.Req.Method, strconv.Itoa(ctx.RespStatusCode)).Observe(float64(duration))
		}()
		ctx.Next()
	}
}
