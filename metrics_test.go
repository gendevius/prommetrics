package prommetrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

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
			got := addLabel(tt.base, tt.key, tt.value)
			assert.Equal(t, tt.want, got)
		})
	}
}
