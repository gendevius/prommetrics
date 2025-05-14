package prommetrics

import (
	"fmt"
	"net/http"
	"strings"
	"time"

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
}

// New creates a new Metrics instance with predefined metrics
func New() *Metrics {
	return &Metrics{
		successRequests: createSuccessCounter(),
		failedRequests:  createFailedCounter(),
		requestDuration: createDurationHistogram(),
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
		[]string{labelVendor, labelEndpoint, labelMethod},
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
		[]string{labelVendor, labelEndpoint, labelMethod},
	)
}

// Middleware creates a wrapper for logging API request metrics
func (m *Metrics) Middleware(vendor string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := newResponseRecorder(w)

			defer func() {
				duration := time.Since(start).Seconds()
				endpoint := extractEndpoint(r.URL.Path)
				method := r.Method
				status := rw.status

				baseLabels := prometheus.Labels{
					labelVendor:   vendor,
					labelEndpoint: endpoint,
					labelMethod:   method,
				}

				m.requestDuration.With(baseLabels).Observe(duration)

				if isSuccess(status) {
					m.successRequests.With(baseLabels).Inc()
				} else {
					m.failedRequests.With(mergeLabels(baseLabels, "code", fmt.Sprint(rw.status))).Inc()
				}
			}()

			next.ServeHTTP(rw, r)
		})
	}
}

func isSuccess(status int) bool {
	return status >= 200 && status < 400
}

func mergeLabels(base prometheus.Labels, key, value string) prometheus.Labels {
	newLabels := make(prometheus.Labels, len(base)+1)
	for k, v := range base {
		newLabels[k] = v
	}
	newLabels[key] = value
	return newLabels
}

func extractEndpoint(path string) string {
	cleanPath := strings.Trim(path, "/")
	if cleanPath == "" {
		return "root"
	}

	parts := strings.Split(cleanPath, "/")
	lastPart := parts[len(parts)-1]

	if lastPart == "" && len(parts) > 1 {
		return parts[len(parts)-2]
	}
	return lastPart
}

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
