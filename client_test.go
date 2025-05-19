package prommetrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRoundTripper struct {
	resp *http.Response
	err  error
}

func (m *mockRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	time.Sleep(5 * time.Millisecond)
	return m.resp, m.err
}

func TestInstrumentedTransport(t *testing.T) {
	t.Parallel()

	metrics := New("test-transport")

	tests := []struct {
		name        string
		mockResp    *http.Response
		mockErr     error
		request     *http.Request
		wantStatus  int
		wantSuccess float64
		wantFailed  float64
		wantLabels  prometheus.Labels
	}{
		{
			name:        "successful GET request",
			mockResp:    &http.Response{StatusCode: http.StatusOK},
			request:     httptest.NewRequest("GET", "http://api.example.com/users", nil),
			wantStatus:  http.StatusOK,
			wantSuccess: 1,
			wantFailed:  0,
			wantLabels: prometheus.Labels{
				labelVendor:   "test-transport",
				labelEndpoint: "http://api.example.com/users",
				labelMethod:   "GET",
				labelCode:     "200",
			},
		},
		{
			name:        "failed POST request (500)",
			mockResp:    &http.Response{StatusCode: http.StatusInternalServerError},
			request:     httptest.NewRequest("POST", "http://api.example.com/error", nil),
			wantStatus:  http.StatusInternalServerError,
			wantSuccess: 0,
			wantFailed:  1,
			wantLabels: prometheus.Labels{
				labelVendor:   "test-transport",
				labelEndpoint: "http://api.example.com/error",
				labelMethod:   "POST",
				labelCode:     "500",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics.successRequests.Reset()
			metrics.failedRequests.Reset()

			transport := &instrumentedTransport{
				next:    &mockRoundTripper{resp: tt.mockResp, err: tt.mockErr},
				metrics: metrics,
			}

			resp, err := transport.RoundTrip(tt.request)

			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			if tt.wantSuccess > 0 {
				val := testutil.ToFloat64(metrics.successRequests.With(tt.wantLabels))
				assert.Equal(t, tt.wantSuccess, val)
			}

			if tt.wantFailed > 0 {
				val := testutil.ToFloat64(metrics.failedRequests.With(tt.wantLabels))
				assert.Equal(t, tt.wantFailed, val)
			}

			_, err = metrics.requestDuration.GetMetricWith(tt.wantLabels)
			assert.NoError(t, err, "histogram should exist")

			metricCount := testutil.CollectAndCount(metrics.requestDuration)
			assert.Greater(t, metricCount, 0, "no metrics collected")
		})
	}
}

func TestInstrumentClient_Wrapping(t *testing.T) {
	metrics := New("test-wrapping")
	client := &http.Client{Transport: http.DefaultTransport}

	instrumented := metrics.InstrumentClient(client)

	wrapped, ok := instrumented.Transport.(*instrumentedTransport)
	require.True(t, ok)
	assert.Equal(t, http.DefaultTransport, wrapped.next)
	assert.Equal(t, metrics, wrapped.metrics)
}

func TestInstrumentClient_NilTransport(t *testing.T) {
	metrics := New("test-nill-transport")
	client := &http.Client{Transport: nil}

	instrumented := metrics.InstrumentClient(client)
	assert.NotNil(t, instrumented.Transport)
}

//func NewTestMetrics() *Metrics {
//	metrics := &Metrics{
//		successRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
//			Namespace: namespace,
//			Subsystem: subsystem,
//			Name:      "success_total",
//			Help:      "Total number of successful vendor API requests",
//		},
//			[]string{labelVendor, labelEndpoint, labelMethod, labelCode}),
//		failedRequests: prometheus.NewCounterVec(
//			prometheus.CounterOpts{
//				Namespace: namespace,
//				Subsystem: subsystem,
//				Name:      "failed_total",
//				Help:      "Total number of failed vendor API requests",
//			},
//			[]string{labelVendor, labelEndpoint, labelMethod, labelCode}),
//		requestDuration: prometheus.NewHistogramVec(
//			prometheus.HistogramOpts{
//				Namespace: namespace,
//				Subsystem: subsystem,
//				Name:      "duration_seconds",
//				Help:      "Duration of vendor API requests",
//				Buckets:   prometheus.DefBuckets,
//			},
//			[]string{labelVendor, labelEndpoint, labelMethod, labelCode}),
//		vendor: "test-vendor",
//	}
//
//	registry := prometheus.NewRegistry()
//	registry.MustRegister(metrics.failedRequests, metrics.successRequests, metrics.requestDuration)
//
//	return &Metrics{
//		successRequests: metrics.successRequests,
//		failedRequests:  metrics.failedRequests,
//		requestDuration: metrics.requestDuration,
//		vendor:          metrics.vendor,
//	}
//}
