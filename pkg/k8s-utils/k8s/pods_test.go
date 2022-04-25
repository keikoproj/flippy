package k8s

import (
	"k8s.io/client-go/kubernetes"
	"testing"
	"time"
)

func TestK8sWrapper_GetLogFromFirstPod(t *testing.T) {
	type args struct {
		clientset       kubernetes.Interface
		namespace       string
		podNameContains string
		containerName   string
		logsFromTime    time.Time
	}

	testArg := args{
		clientset:       BuildTestK8SClientSet().K8sClientSet,
		namespace:       "Sample",
		podNameContains: "Sample",
		containerName:   "Sample",
		logsFromTime:    time.Now(),
	}

	//testClientSet := BuildTestK8SClientSet().K8sClientSet

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"Sample", testArg, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k8 := K8sWrapper{}
			got, err := k8.GetLogFromFirstPod(tt.args.clientset, tt.args.namespace, tt.args.podNameContains, tt.args.containerName, tt.args.logsFromTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLogFromFirstPod() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetLogFromFirstPod() got = %v, want %v", got, tt.want)
			}
		})
	}
}
