package lsego

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

var (
	requestLabels = []string{"status", "endpoint", "method"}

	uptime = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "process_uptime_seconds",
			Help: "service process uptime seconds.",
		}, nil,
	)

	reqCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_server_request_count_total",
			Help: "Total number of HTTP requests made.",
		}, requestLabels,
	)

	reqDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_server_request_duration_seconds",
			Help: "HTTP request latencies in seconds.",
		}, requestLabels,
	)

	reqSizeBytes = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_server_request_size_bytes",
			Help: "HTTP request sizes in bytes.",
		}, requestLabels,
	)

	respSizeBytes = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_server_response_size_bytes",
			Help: "HTTP response sizes in bytes.",
		}, requestLabels,
	)
)

// init registers the prometheus metrics
func init() {
	prometheus.MustRegister(collectors.NewBuildInfoCollector())
	prometheus.MustRegister(uptime, reqCount, reqDuration, reqSizeBytes, respSizeBytes)
	go recordUptime()
}

// recordUptime increases service uptime per second.
func recordUptime() {
	for range time.Tick(time.Second) {
		uptime.WithLabelValues().Inc()
	}
}

func metricMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := newMetricRw(w)
		start := time.Now()
		endpoint := r.URL.Path
		method := r.Method
		next.ServeHTTP(rw, r)

		status := fmt.Sprintf("%d", rw.Code)
		lvs := []string{status, endpoint, method}
		respSize := rw.Written
		if respSize < 0 {
			respSize = 0
		}

		reqCount.WithLabelValues(lvs...).Inc()
		reqDuration.WithLabelValues(lvs...).Observe(time.Since(start).Seconds())
		reqSizeBytes.WithLabelValues(lvs...).Observe(float64(r.ContentLength))
		respSizeBytes.WithLabelValues(lvs...).Observe(float64(respSize))
	})
}
