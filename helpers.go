package prommetrics

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gofrs/uuid/v5"
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
		return "/"
	}

	var builder strings.Builder
	parts := strings.Split(cleanPath, "/")

	for i, part := range parts {
		if isID(part) {
			builder.WriteString("id")
		} else {
			builder.WriteString(part)
		}

		if i < len(parts)-1 {
			builder.WriteString("/")
		}
	}

	return builder.String()
}

func isID(s string) bool {
	return isUuid(s) || isInt(s)
}

func isUuid(s string) bool {
	_, err := uuid.FromString(s)
	return err == nil
}

func isInt(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}
