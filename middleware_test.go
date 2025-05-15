package prommetrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware(t *testing.T) {
	t.Parallel()

	metrics := NewTestMetrics()

	tests := []struct {
		name             string
		handler          http.HandlerFunc
		request          *http.Request
		wantStatus       int
		wantSuccess      float64
		wantFailed       float64
		wantLabels       prometheus.Labels
		wantFailedLabels prometheus.Labels
	}{
		{
			name: "successful GET request",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			request:     httptest.NewRequest("GET", "/api/users", nil),
			wantStatus:  http.StatusOK,
			wantSuccess: 1,
			wantFailed:  0,
			wantLabels: prometheus.Labels{
				labelVendor:   "test-vendor",
				labelEndpoint: "users",
				labelMethod:   "GET",
			},
		},
		{
			name: "successful POST request",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusCreated)
			},
			request:     httptest.NewRequest("POST", "/api/products", nil),
			wantStatus:  http.StatusCreated,
			wantSuccess: 1,
			wantFailed:  0,
			wantLabels: prometheus.Labels{
				labelVendor:   "test-vendor",
				labelEndpoint: "products",
				labelMethod:   "POST",
			},
		},
		{
			name: "failed request (500)",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			request:     httptest.NewRequest("GET", "/api/error", nil),
			wantStatus:  http.StatusInternalServerError,
			wantSuccess: 0,
			wantFailed:  1,
			wantLabels: prometheus.Labels{
				labelVendor:   "test-vendor",
				labelEndpoint: "error",
				labelMethod:   "GET",
			},
			wantFailedLabels: prometheus.Labels{
				labelVendor:   "test-vendor",
				labelEndpoint: "error",
				labelMethod:   "GET",
				labelCode:     "500",
			},
		},
		{
			name: "not found (404)",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			request:     httptest.NewRequest("GET", "/api/notfound", nil),
			wantStatus:  http.StatusNotFound,
			wantSuccess: 0,
			wantFailed:  1,
			wantLabels: prometheus.Labels{
				labelVendor:   "test-vendor",
				labelEndpoint: "notfound",
				labelMethod:   "GET",
			},
			wantFailedLabels: prometheus.Labels{
				labelVendor:   "test-vendor",
				labelEndpoint: "notfound",
				labelMethod:   "GET",
				labelCode:     "404",
			},
		},
		{
			name: "root path",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			request:     httptest.NewRequest("GET", "/", nil),
			wantStatus:  http.StatusOK,
			wantSuccess: 1,
			wantFailed:  0,
			wantLabels: prometheus.Labels{
				labelVendor:   "test-vendor",
				labelEndpoint: "root",
				labelMethod:   "GET",
			},
		},
		{
			name: "long nested path",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			request:     httptest.NewRequest("GET", "/api/v1/users/12345/profile", nil),
			wantStatus:  http.StatusOK,
			wantSuccess: 1,
			wantFailed:  0,
			wantLabels: prometheus.Labels{
				labelVendor:   "test-vendor",
				labelEndpoint: "profile",
				labelMethod:   "GET",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics.successRequests.Reset()
			metrics.failedRequests.Reset()

			rr := httptest.NewRecorder()
			metrics.Middleware()(tt.handler).ServeHTTP(rr, tt.request)

			if status := rr.Code; status != tt.wantStatus {
				t.Errorf("wrong status code: got %v want %v", status, tt.wantStatus)
			}

			if tt.wantSuccess > 0 {
				val := testutil.ToFloat64(metrics.successRequests.With(tt.wantLabels))
				if val != tt.wantSuccess {
					t.Errorf("wrong success count: got %v want %v", val, tt.wantSuccess)
				}
			}

			if tt.wantFailed > 0 {
				val := testutil.ToFloat64(metrics.failedRequests.With(tt.wantFailedLabels))
				if val != tt.wantFailed {
					t.Errorf("wrong failed count: got %v want %v", val, tt.wantFailed)
				}
			}

			_, err := metrics.requestDuration.GetMetricWith(tt.wantLabels)
			assert.NoError(t, err, "histogram metric should exist")

			metricCount := testutil.CollectAndCount(metrics.requestDuration)
			assert.Greater(t, metricCount, 0, "histogram should have observations")
		})
	}
}
