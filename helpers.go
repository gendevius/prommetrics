package prommetrics

import (
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

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

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}
