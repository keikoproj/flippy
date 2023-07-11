package common

import (
	"testing"
	"time"
)

func TestGetReconcileDuration(t *testing.T) {
	type args struct {
		flagReconcilerTime string
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{"empty", args{flagReconcilerTime: ""}, 10 * time.Hour},
		{"Invalid 1 hr", args{flagReconcilerTime: "1hr"}, 10 * time.Hour},
		{"Invalid 1 hr", args{flagReconcilerTime: "1h"}, 1 * time.Hour},
		{"Invalid Input", args{flagReconcilerTime: "random"}, 10 * time.Hour},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetReconcileDuration(tt.args.flagReconcilerTime); got != tt.want {
				t.Errorf("GetReconcileDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}
