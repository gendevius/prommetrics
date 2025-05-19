package prommetrics

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// instrumentedTransport wraps a http.RoundTripper with metrics collection
type instrumentedTransport struct {
	next    http.RoundTripper
	metrics *Metrics
}

// RoundTrip implements http.RoundTripper interface
func (t *instrumentedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	endpoint := extractEndpoint(req.URL.String())
	method := req.Method

	baseLabels := prometheus.Labels{
		labelVendor:   t.metrics.vendor,
		labelEndpoint: endpoint,
		labelMethod:   method,
	}

	resp, err := t.next.RoundTrip(req)
	if err != nil {
		if resp.StatusCode != 0 {
			t.metrics.failedRequests.With(addLabel(baseLabels, labelCode, fmt.Sprint(resp.StatusCode))).Inc()
			return nil, err
		}
		t.metrics.failedRequests.With(addLabel(baseLabels, labelCode, fmt.Sprint(http.StatusInternalServerError))).Inc()
		return nil, err
	}
	defer func() {
		duration := time.Since(start).Seconds()

		t.metrics.requestDuration.With(addLabel(baseLabels, labelCode, fmt.Sprint(resp.StatusCode))).Observe(duration)

		if isSuccessStatus(resp.StatusCode) {
			t.metrics.successRequests.With(addLabel(baseLabels, labelCode, fmt.Sprint(resp.StatusCode))).Inc()
			return
		}
		t.metrics.failedRequests.With(addLabel(baseLabels, labelCode, fmt.Sprint(resp.StatusCode))).Inc()
	}()

	return resp, nil
}

// InstrumentClient instruments an existing http.Client with metrics collection
func (m *Metrics) InstrumentClient(client *http.Client) http.Client {
	if client == nil {
		client = &http.Client{}
	}

	originalTransport := client.Transport
	if originalTransport == nil {
		originalTransport = http.DefaultTransport
	}

	client.Transport = &instrumentedTransport{
		next:    originalTransport,
		metrics: m,
	}

	return *client
}
