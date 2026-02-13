package observability

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	registerMetricsOnce sync.Once

	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "texa",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "texa",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "Duration of HTTP requests in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
	saleCreateTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "texa",
		Subsystem: "business",
		Name:      "sale_create_total",
		Help:      "Total sales created.",
	})
	refundCreateTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "texa",
		Subsystem: "business",
		Name:      "refund_create_total",
		Help:      "Total refunds created.",
	})
	transferCreateTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "texa",
		Subsystem: "business",
		Name:      "transfer_create_total",
		Help:      "Total branch transfers created.",
	})
)

func RegisterMetrics() {
	registerMetricsOnce.Do(func() {
		prometheus.MustRegister(httpRequestsTotal, httpRequestDuration, saleCreateTotal, refundCreateTotal, transferCreateTotal)
	})
}

func ObserveHTTP(method, path, status string, durationSeconds float64) {
	RegisterMetrics()
	httpRequestsTotal.WithLabelValues(method, path, status).Inc()
	httpRequestDuration.WithLabelValues(method, path).Observe(durationSeconds)
}

func IncSaleCreate() {
	RegisterMetrics()
	saleCreateTotal.Inc()
}

func IncRefundCreate() {
	RegisterMetrics()
	refundCreateTotal.Inc()
}

func IncTransferCreate() {
	RegisterMetrics()
	transferCreateTotal.Inc()
}
