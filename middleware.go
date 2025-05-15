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

				m.requestDuration.With(mergeLabels(baseLabels, labelCode, fmt.Sprint(rw.status))).Observe(duration)

				if isSuccess(status) {
					m.successRequests.With(mergeLabels(baseLabels, labelCode, fmt.Sprint(rw.status))).Inc()
				} else {
					m.failedRequests.With(mergeLabels(baseLabels, labelCode, fmt.Sprint(rw.status))).Inc()
				}
			}()

			next.ServeHTTP(rw, r)
		})
	}
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
