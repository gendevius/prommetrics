package prommetrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	namespace = "vendor_api"
	subsystem = "requests"

	labelVendor   = "vendor"
	labelEndpoint = "endpoint"
	labelMethod   = "method"
	labelCode     = "code"
)

// Metrics introduces a set of metrics for monitoring API requests
type Metrics struct {
	successRequests *prometheus.CounterVec
	failedRequests  *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	vendor          string
}

// New creates a new Metrics instance with predefined metrics
func New(vendor string) *Metrics {
	return &Metrics{
		successRequests: createSuccessCounter(),
		failedRequests:  createFailedCounter(),
		requestDuration: createDurationHistogram(),
		vendor:          vendor,
	}
}

func createSuccessCounter() *prometheus.CounterVec {
	return promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "success_total",
			Help:      "Total number of successful vendor API requests",
		},
		[]string{labelVendor, labelEndpoint, labelMethod, labelCode},
	)
}

func createFailedCounter() *prometheus.CounterVec {
	return promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "failed_total",
			Help:      "Total number of failed vendor API requests",
		},
		[]string{labelVendor, labelEndpoint, labelMethod, labelCode},
	)
}

func createDurationHistogram() *prometheus.HistogramVec {
	return promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "duration_seconds",
			Help:      "Duration of vendor API requests",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{labelVendor, labelEndpoint, labelMethod, labelCode},
	)
}

func addLabel(base prometheus.Labels, key, value string) prometheus.Labels {
	newLabels := make(prometheus.Labels, len(base)+1)
	for k, v := range base {
		newLabels[k] = v
	}
	newLabels[key] = value
	return newLabels
}
