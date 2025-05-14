package prommetrics

import (
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	namespace = "vendor_api"
	subsystem = "requests"
)

type Metrics struct {
	successRequests *prometheus.CounterVec
	failedRequests  *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

func New() *Metrics {
	return &Metrics{
		successRequests: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "success_total",
				Help:      "Total number of successful vendor API requests",
			},
			[]string{"vendor", "endpoint", "method"},
		),

		failedRequests: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "failed_total",
				Help:      "Total number of failed vendor API requests",
			},
			[]string{"vendor", "endpoint", "method", "code"},
		),

		requestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "duration_seconds",
				Help:      "Duration of vendor API requests",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"vendor", "endpoint", "method"},
		),
	}
}

func (m *Metrics) Middleware(vendor string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseRecorder{ResponseWriter: w}

			next.ServeHTTP(rw, r)

			duration := time.Since(start).Seconds()
			endpoint := extractEndpoint(r.URL.Path)
			method := r.Method
			status := rw.status

			labels := prometheus.Labels{
				"vendor":   vendor,
				"endpoint": endpoint,
				"method":   method,
			}

			m.requestDuration.With(labels).Observe(duration)

			if status >= 200 && status < 400 {
				m.successRequests.With(labels).Inc()
			} else {
				failedLabels := prometheus.Labels{
					"vendor":   vendor,
					"endpoint": endpoint,
					"method":   method,
					"code":     http.StatusText(status),
				}
				m.failedRequests.With(failedLabels).Inc()
			}
		})
	}
}

func extractEndpoint(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return "/"
	}
	return parts[len(parts)-1]
}

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
