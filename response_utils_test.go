package prommetrics

import (
	"net/http"
	"testing"

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
			assert.Equal(t, tt.want, isSuccessStatus(tt.status))
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
			want: "/",
		},
		{
			name: "simple path",
			path: "/api/test",
			want: "api/test",
		},
		{
			name: "nested path",
			path: "/api/v1/users",
			want: "api/v1/users",
		},
		{
			name: "path with trailing slash",
			path: "/api/v1/users/",
			want: "api/v1/users",
		},
		{
			name: "empty path",
			path: "",
			want: "/",
		},
		{
			name: "long path",
			path: "/very/long/path/to/resource",
			want: "very/long/path/to/resource",
		},
		{
			name: "path with int",
			path: "/very/long/path/1/resource",
			want: "very/long/path/<int>/resource",
		},
		{
			name: "path with uuid",
			path: "/very/long/path/1ca0d4c6-796e-4a1c-a6a0-95fb0f4033b6/resource",
			want: "very/long/path/<uuid>/resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, extractEndpoint(tt.path))
		})
	}
}
