package k8s

import (
	"errors"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
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

// TestDeletePod_ComprehensiveCoverage tests all paths in DeletePod function
func TestDeletePod_ComprehensiveCoverage(t *testing.T) {
	tests := []struct {
		name           string
		setupClientset func() kubernetes.Interface
		namespace      string
		podName        string
		expectError    bool
		description    string
	}{
		{
			name: "SuccessfulDeletionFirstAttempt",
			setupClientset: func() kubernetes.Interface {
				// Create fake clientset with a pod
				return fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "test-namespace",
					},
				})
			},
			namespace:   "test-namespace",
			podName:     "test-pod",
			expectError: false,
			description: "Should successfully delete pod on first attempt",
		},
		{
			name: "SuccessfulDeletionSecondAttempt",
			setupClientset: func() kubernetes.Interface {
				fakeClient := fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "test-namespace",
					},
				})

				// Mock first delete to fail, second to succeed
				callCount := 0
				fakeClient.PrependReactor("delete", "pods", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					callCount++
					if callCount == 1 {
						// First call fails
						return true, nil, errors.New("temporary failure")
					}
					// Second call succeeds
					return false, nil, nil // Let default behavior handle success
				})

				return fakeClient
			},
			namespace:   "test-namespace",
			podName:     "test-pod",
			expectError: false,
			description: "Should successfully delete pod on second attempt after first fails",
		},
		{
			name: "FailureBothAttempts",
			setupClientset: func() kubernetes.Interface {
				fakeClient := fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "test-namespace",
					},
				})

				// Mock both deletes to fail
				fakeClient.PrependReactor("delete", "pods", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("persistent failure")
				})

				return fakeClient
			},
			namespace:   "test-namespace",
			podName:     "test-pod",
			expectError: true,
			description: "Should return error when both delete attempts fail",
		},
		{
			name: "NonExistentPod",
			setupClientset: func() kubernetes.Interface {
				// Empty clientset - no pods
				return fake.NewSimpleClientset()
			},
			namespace:   "test-namespace",
			podName:     "non-existent-pod",
			expectError: true,
			description: "Should return error when trying to delete non-existent pod",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientset := tt.setupClientset()
			wrapper := K8sWrapper{}

			err := wrapper.DeletePod(clientset, tt.namespace, tt.podName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none: %s", tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got %v: %s", err, tt.description)
				}
			}
		})
	}
}

// TestGetRunningPods_ComprehensiveCoverage tests all paths in GetRunningPods function
func TestGetRunningPods_ComprehensiveCoverage(t *testing.T) {
	tests := []struct {
		name            string
		setupClientset  func() kubernetes.Interface
		namespace       string
		podNameContains string
		expectError     bool
		expectedPods    int
		description     string
	}{
		{
			name: "SuccessfulPodRetrieval",
			setupClientset: func() kubernetes.Interface {
				return fake.NewSimpleClientset(
					&corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "running-pod-1",
							Namespace: "test-namespace",
						},
						Status: corev1.PodStatus{
							Phase: corev1.PodRunning,
						},
					},
					&corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "running-pod-2",
							Namespace: "test-namespace",
						},
						Status: corev1.PodStatus{
							Phase: corev1.PodRunning,
						},
					},
					&corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pending-pod",
							Namespace: "test-namespace",
						},
						Status: corev1.PodStatus{
							Phase: corev1.PodPending,
						},
					},
				)
			},
			namespace:       "test-namespace",
			podNameContains: "running-pod",
			expectError:     false,
			expectedPods:    2,
			description:     "Should return only running pods that match name filter",
		},
		{
			name: "NoMatchingPods",
			setupClientset: func() kubernetes.Interface {
				return fake.NewSimpleClientset(
					&corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "other-pod",
							Namespace: "test-namespace",
						},
						Status: corev1.PodStatus{
							Phase: corev1.PodRunning,
						},
					},
				)
			},
			namespace:       "test-namespace",
			podNameContains: "running-pod",
			expectError:     false,
			expectedPods:    0,
			description:     "Should return empty list when no pods match filter",
		},
		{
			name: "EmptyNamespace",
			setupClientset: func() kubernetes.Interface {
				return fake.NewSimpleClientset()
			},
			namespace:       "test-namespace",
			podNameContains: "any-pod",
			expectError:     false,
			expectedPods:    0,
			description:     "Should return empty list when namespace has no pods",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientset := tt.setupClientset()
			wrapper := K8sWrapper{}

			pods, err := wrapper.GetRunningPods(clientset, tt.namespace, tt.podNameContains)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none: %s", tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got %v: %s", err, tt.description)
				}
				if len(pods) != tt.expectedPods {
					t.Errorf("Expected %d pods but got %d: %s", tt.expectedPods, len(pods), tt.description)
				}
			}
		})
	}
}
