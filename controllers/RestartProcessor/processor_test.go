package RestartProcessor

import "testing"

func TestIsRestartGood(t *testing.T) {
	type args struct {
		output string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"10 Restart", args{output: "rollout to finish:10"}, true},
		{"0 Restart", args{output: "rollout to finish:0"}, false},
		{"1 Restart", args{output: "rollout to finish:1"}, true},
		{"Empty Output", args{output: ""}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRestartGood(tt.args.output); got != tt.want {
				t.Errorf("IsRestartGood() = %v, want %v", got, tt.want)
			}
		})
	}
}
