package cmdutil

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	metricCmdExecTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cmd_exec_total",
			Help: "total number of cmd exec",
		},
		[]string{"command", "param"},
	)

	metricCmdExecLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cmd_exec_latency",
			Help:    "the latency of cmd exec",
			Buckets: []float64{1.0, 3.0, 5.0, 7.0, 9.0, 12.0, 15.0, 18.0, 30.0, 50.0, 100.0}, // for second
		},
		[]string{"command", "param"},
	)
)

func init() {
	prometheus.MustRegister(metricCmdExecTotal)
	prometheus.MustRegister(metricCmdExecLatency)
}
