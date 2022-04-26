package k8s

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
	"time"
)

func TestK8sPodWrapper(t *testing.T) {
	type args struct {
		clientset       kubernetes.Interface
		namespace       string
		podNameContains string
		containerName   string
		logsFromTime    time.Time
	}

	fakeClientSet := fake.NewSimpleClientset(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "TestPodName",
				Namespace:   "TestNamespace",
				Annotations: map[string]string{},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		}, &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "TestPodName2",
				Namespace:   "TestNamespace",
				Annotations: map[string]string{},
			},
		})

	testArg := args{
		clientset:       fakeClientSet,
		namespace:       "TestNamespace",
		podNameContains: "TestPodName",
		containerName:   "TestContainerName",
		logsFromTime:    time.Now(),
	}

	nonExistingPodArg := args{
		clientset:       fakeClientSet,
		namespace:       "TestNamespace",
		podNameContains: "TestPodName3",
		containerName:   "TestContainerName",
		logsFromTime:    time.Now(),
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"TestGetLogFromFirstPod", testArg, "fake logs", false},
		{"TestGetLogFromFirstPodForNonExistingPod", nonExistingPodArg, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k8 := K8sWrapper{}

			podLog, err := k8.GetLogFromFirstPod(tt.args.clientset, tt.args.namespace, tt.args.podNameContains, tt.args.containerName, tt.args.logsFromTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLogFromFirstPod() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if podLog != tt.want {
				t.Errorf("GetLogFromFirstPod() got = %v, want %v", podLog, tt.want)
			}
		})
	}

	deletePodTests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"TestDeletePod", testArg, "", false},
		{"TestDeletePodForNonExistingPod", nonExistingPodArg, "", false},
	}

	for _, tt := range deletePodTests {
		t.Run(tt.name, func(t *testing.T) {
			k8 := K8sWrapper{}

			err := k8.DeletePods(tt.args.clientset, tt.args.namespace, tt.args.podNameContains)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLogFromFirstPod() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
