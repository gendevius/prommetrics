package prommetrics

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Middleware creates a wrapper for logging API request metrics
func (m *Metrics) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := newResponseRecorder(w)

			defer func() {
				duration := time.Since(start).Seconds()
				endpoint := extractEndpoint(r.RequestURI)
				method := r.Method
				status := rw.status

				baseLabels := prometheus.Labels{
					labelVendor:   m.vendor,
					labelEndpoint: endpoint,
					labelMethod:   method,
				}

				m.requestDuration.With(addLabel(baseLabels, labelCode, fmt.Sprint(rw.status))).Observe(duration)

				if isSuccessStatus(status) {
					m.successRequests.With(addLabel(baseLabels, labelCode, fmt.Sprint(rw.status))).Inc()
					return
				}
				m.failedRequests.With(addLabel(baseLabels, labelCode, fmt.Sprint(rw.status))).Inc()

			}()

			next.ServeHTTP(rw, r)
		})
	}
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
