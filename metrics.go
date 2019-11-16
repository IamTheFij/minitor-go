package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	// "github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

// TODO: Not sure if this is the best way to handle. A global instance for
// metrics isn't bad, but it might be nice to curry versions of the metrics
// for each monitor. Especially since every monitor has it's own. Perhaps
// another new function that essentially curries each metric for a given
// monitor name would do. This could be run when validating monitors and
// initializing alert templates.

// MinitorMetrics contains all counters and metrics that Minitor will need to access
type MinitorMetrics struct {
	alertCount    *prometheus.CounterVec
	checkCount    *prometheus.CounterVec
	monitorStatus *prometheus.GaugeVec
}

// NewMetrics creates and initializes all metrics
func NewMetrics() *MinitorMetrics {
	// Initialize all metrics
	metrics := &MinitorMetrics{
		alertCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "minitor_alert_total",
				Help: "Number of Minitor alerts",
			},
			[]string{"alert", "monitor"},
		),
		checkCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "minitor_check_total",
				Help: "Number of Minitor checks",
			},
			[]string{"monitor", "status", "is_alert"},
		),
		monitorStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "minitor_monitor_up_count",
				Help: "Status of currently responsive monitors",
			},
			[]string{"monitor"},
		),
	}

	// Register newly created metrics
	prometheus.MustRegister(metrics.alertCount)
	prometheus.MustRegister(metrics.checkCount)
	prometheus.MustRegister(metrics.monitorStatus)

	return metrics
}

// SetMonitorStatus sets the current status of Monitor
func (metrics *MinitorMetrics) SetMonitorStatus(monitor string, isUp bool) {
	val := 0.0
	if isUp {
		val = 1.0
	}
	metrics.monitorStatus.With(prometheus.Labels{"monitor": monitor}).Set(val)
}

// CountCheck counts the result of a particular Monitor check
func (metrics *MinitorMetrics) CountCheck(monitor string, isSuccess bool, isAlert bool) {
	status := "failure"
	if isSuccess {
		status = "success"
	}

	alertVal := "false"
	if isAlert {
		alertVal = "true"
	}

	metrics.checkCount.With(
		prometheus.Labels{"monitor": monitor, "status": status, "is_alert": alertVal},
	).Inc()
}

// CountAlert counts an alert
func (metrics *MinitorMetrics) CountAlert(monitor string, alert string) {
	metrics.alertCount.With(
		prometheus.Labels{
			"alert":   alert,
			"monitor": monitor,
		},
	).Inc()
}

// ServeMetrics starts an http server with a Prometheus metrics handler
func ServeMetrics() {
	http.Handle("/metrics", promhttp.Handler())
	host := fmt.Sprintf(":%d", MetricsPort)
	_ = http.ListenAndServe(host, nil)
}