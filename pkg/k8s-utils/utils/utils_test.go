package utils

import (
	"github.com/keikoproj/flippy/pkg/common"
	"github.com/tj/assert"
	"testing"
)

func TestIsStringMapSubset(t *testing.T) {

	masterMap := make(map[string]string)
	masterMap["test1"] = "test"
	masterMap["test2"] = ""
	masterMap["test3"] = "true"
	masterMap["test4"] = "false"

	subsetMap := make(map[string]string)
	subsetMap["test1"] = "test"

	tests := []struct {
		name string
		args string
		want bool
	}{
		{"No addition to label", "empty", true},
		{"Addition to label to ignore flippy with empty value", "test2", true},
		{"Addition to label to ignore flippy", "test3", false},
		{"Addition to label to ignore flippy with default value", "test4", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args != "empty" {
				common.IgnoreMetadataKey = tt.args
			}
			if got := IsStringMapSubset(masterMap, subsetMap); got != tt.want {
				t.Errorf("IsStringMapSubset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsStringMapSubsetNegative(t *testing.T) {

	masterMap := make(map[string]string)
	masterMap["test1"] = "test"
	masterMap["test2"] = ""
	masterMap["test3"] = "true"
	masterMap["test4"] = "false"

	subsetMap := make(map[string]string)
	subsetMap["test5"] = "test"

	got := IsStringMapSubset(masterMap, subsetMap)
	assert.Equal(t, false, got)
}
