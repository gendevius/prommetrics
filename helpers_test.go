package prommetrics

import (
	"net/http"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestIsSuccess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status int
		want   bool
	}{
		{
			name:   "status 200",
			status: http.StatusOK,
			want:   true,
		},
		{
			name:   "status 201",
			status: http.StatusCreated,
			want:   true,
		},
		{
			name:   "status 300",
			status: http.StatusMultipleChoices,
			want:   true,
		},
		{
			name:   "status 400",
			status: http.StatusBadRequest,
			want:   false,
		},
		{
			name:   "status 500",
			status: http.StatusInternalServerError,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, isSuccess(tt.status))
		})
	}
}

func TestMergeLabels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		base  prometheus.Labels
		key   string
		value string
		want  prometheus.Labels
	}{
		{
			name:  "add to empty labels",
			base:  prometheus.Labels{},
			key:   "new_key",
			value: "new_value",
			want:  prometheus.Labels{"new_key": "new_value"},
		},
		{
			name:  "add to existing labels",
			base:  prometheus.Labels{"exist": "value"},
			key:   "new_key",
			value: "new_value",
			want:  prometheus.Labels{"exist": "value", "new_key": "new_value"},
		},
		{
			name:  "override existing key",
			base:  prometheus.Labels{"key": "old"},
			key:   "key",
			value: "new",
			want:  prometheus.Labels{"key": "new"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := mergeLabels(tt.base, tt.key, tt.value)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExtractEndpoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "root path",
			path: "/",
			want: "root",
		},
		{
			name: "simple path",
			path: "/api/test",
			want: "test",
		},
		{
			name: "nested path",
			path: "/api/v1/users",
			want: "users",
		},
		{
			name: "path with trailing slash",
			path: "/api/v1/users/",
			want: "users",
		},
		{
			name: "empty path",
			path: "",
			want: "root",
		},
		{
			name: "long path",
			path: "/very/long/path/to/resource",
			want: "resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, extractEndpoint(tt.path))
		})
	}
}
