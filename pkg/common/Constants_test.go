package common

import "testing"

func TestParseStaggerNamespacesEnabled(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"missing", "", false},
		{"true", "true", true},
		{"TRUE", "TRUE", true},
		{"1", "1", true},
		{"false", "false", false},
		{"0", "0", false},
		{"invalid word", "yes", false},
		{"garbage", "enabled", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseStaggerNamespacesEnabled(tt.value); got != tt.want {
				t.Errorf("ParseStaggerNamespacesEnabled(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}
