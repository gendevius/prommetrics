package prommetrics

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// instrumentedTransport wraps an http.RoundTripper with metrics collection
type instrumentedTransport struct {
	next    http.RoundTripper
	metrics *Metrics
}

// RoundTrip implements http.RoundTripper interface
func (t *instrumentedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	endpoint := extractEndpoint(req.URL.Path)
	method := req.Method

	baseLabels := prometheus.Labels{
		labelEndpoint: endpoint,
		labelMethod:   method,
	}

	resp, err := t.next.RoundTrip(req)
	duration := time.Since(start).Seconds()

	t.metrics.requestDuration.With(baseLabels).Observe(duration)

	if err != nil {
		t.metrics.failedRequests.With(mergeLabels(baseLabels, labelCode, resp.Status)).Inc()
		return nil, err
	}

	if isSuccess(resp.StatusCode) {
		t.metrics.successRequests.With(baseLabels).Inc()
	} else {
		t.metrics.failedRequests.With(mergeLabels(baseLabels, labelCode, fmt.Sprint(resp.StatusCode))).Inc()
	}

	return resp, nil
}

// InstrumentClient instruments an existing http.Client with metrics collection
func (m *Metrics) InstrumentClient(client *http.Client) *http.Client {
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

	return client
}
