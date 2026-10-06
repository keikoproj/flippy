package RestartProcessor

import (
	"reflect"
	"strings"
	"testing"
	"time"

	argov1alpha1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	argo "github.com/argoproj/argo-rollouts/pkg/client/clientset/versioned"
	crdv1 "github.com/keikoproj/flippy/api/v1"
	"github.com/keikoproj/flippy/pkg/common"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/kubernetes"
)

// TestIsRestartGood tests the core string parsing logic for restart status
func TestIsRestartGood(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   bool
	}{
		// Core success patterns
		{"successfully rolled out", "successfully rolled out", true},
		{"updated replicas available", "updated replicas are available", true},
		{"healthy status", "Healthy", true},

		// Rollout progress parsing (critical edge cases)
		{"rollout with active pods", "rollout to finish: 5 of 10 updated replicas", true},
		{"rollout with zero pods", "rollout to finish: 0 of 10 updated replicas", false},
		{"invalid number format", "rollout to finish: abc of 10 updated replicas", false},

		// Edge cases that could cause production issues
		{"empty string", "", false},
		{"whitespace handling", "  successfully rolled out  ", true},
		{"multiline output", "line1\nsuccessfully rolled out\nline3", true},
		{"partial match edge case", "not successfully rolled out", true}, // Contains success string

		// Failure scenarios
		{"failure message", "deployment rollout failed", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRestartGood(tt.output); got != tt.want {
				t.Errorf("IsRestartGood() = %v, want %v for input: %q", got, tt.want, tt.output)
			}
		})
	}
}

// Complete MockK8sAPI implementation that satisfies the k8s.K8sAPI interface
type MockK8sAPI struct {
	// RestartProcessor specific methods
	RolloutRestartDeploymentFunc   func(string, string, string) (string, error)
	RolloutDeploymentStatusFunc    func(string, string, string) (string, error)
	RolloutRestartArgoRolloutsFunc func(string, string, string) (string, error)
	RolloutArogRolloutStatusFunc   func(string, string, string) (string, error)

	// Call tracking for verification
	RestartDeploymentCalls []MockCall
	DeploymentStatusCalls  []MockCall
	RestartRolloutCalls    []MockCall
	RolloutStatusCalls     []MockCall

	// Sequence records every restart/status call in invocation order as "<op>:<ns>/<name>"
	// so tests can assert ordering and gating across namespaces.
	Sequence []string
}

type MockCall struct {
	Kubeconfig string
	Namespace  string
	ObjectName string
}

// RestartProcessor specific methods
func (m *MockK8sAPI) RolloutRestartDeployment(kubeconfigpath, namespace, deploymentName string) (string, error) {
	m.RestartDeploymentCalls = append(m.RestartDeploymentCalls, MockCall{kubeconfigpath, namespace, deploymentName})
	m.Sequence = append(m.Sequence, "restart:"+namespace+"/"+deploymentName)
	if m.RolloutRestartDeploymentFunc != nil {
		return m.RolloutRestartDeploymentFunc(kubeconfigpath, namespace, deploymentName)
	}
	return "deployment.apps/" + deploymentName + " restarted", nil
}

func (m *MockK8sAPI) RolloutDeploymentStatus(kubeconfigpath, namespace, deploymentName string) (string, error) {
	m.DeploymentStatusCalls = append(m.DeploymentStatusCalls, MockCall{kubeconfigpath, namespace, deploymentName})
	m.Sequence = append(m.Sequence, "status:"+namespace+"/"+deploymentName)
	if m.RolloutDeploymentStatusFunc != nil {
		return m.RolloutDeploymentStatusFunc(kubeconfigpath, namespace, deploymentName)
	}
	return "deployment \"" + deploymentName + "\" successfully rolled out", nil
}

func (m *MockK8sAPI) RolloutRestartArgoRollouts(kubeconfigpath, namespace, argoRolloutName string) (string, error) {
	m.RestartRolloutCalls = append(m.RestartRolloutCalls, MockCall{kubeconfigpath, namespace, argoRolloutName})
	m.Sequence = append(m.Sequence, "restart:"+namespace+"/"+argoRolloutName)
	if m.RolloutRestartArgoRolloutsFunc != nil {
		return m.RolloutRestartArgoRolloutsFunc(kubeconfigpath, namespace, argoRolloutName)
	}
	return "rollout.argoproj.io/" + argoRolloutName + " restarted", nil
}

func (m *MockK8sAPI) RolloutArogRolloutStatus(kubeconfigpath, namespace, argoRolloutName string) (string, error) {
	m.RolloutStatusCalls = append(m.RolloutStatusCalls, MockCall{kubeconfigpath, namespace, argoRolloutName})
	m.Sequence = append(m.Sequence, "status:"+namespace+"/"+argoRolloutName)
	if m.RolloutArogRolloutStatusFunc != nil {
		return m.RolloutArogRolloutStatusFunc(kubeconfigpath, namespace, argoRolloutName)
	}
	return "Healthy", nil
}

// Stub implementations for all other k8s.K8sAPI interface methods
func (m *MockK8sAPI) PatchResource(kubeconfigpath, namespace, resource, resourceName, patchJson string) (string, error) {
	return "", nil
}
func (m *MockK8sAPI) ScaleDeployment(kubeconfigpath, namespace, deploymentname string, scale int) (string, error) {
	return "", nil
}
func (m *MockK8sAPI) ApplyYaml(kubeconfigpath, yamlFilePath string) (string, error) {
	return "", nil
}
func (m *MockK8sAPI) DeleteYaml(kubeconfigpath, yamlFilePath string) (string, error) {
	return "", nil
}
func (m *MockK8sAPI) GetServiceEntries(kubeconfigpath, namespace string) (map[string]string, error) {
	return nil, nil
}
func (m *MockK8sAPI) RestartContainer(kubeconfigpath, namespace, podname, containerName string) (string, error) {
	return "", nil
}
func (m *MockK8sAPI) ExecuteKubectlCommand(cmdParameter []string) (string, error) {
	return "", nil
}

// Pod methods
func (m *MockK8sAPI) GetRunningPods(clientset kubernetes.Interface, namespace, podNameContains string) ([]corev1.Pod, error) {
	return nil, nil
}
func (m *MockK8sAPI) GetAllPodsInNamespace(clientset kubernetes.Interface, namespace string) ([]corev1.Pod, error) {
	return nil, nil
}
func (m *MockK8sAPI) DeletePod(clientset kubernetes.Interface, namespace, podName string) error {
	return nil
}
func (m *MockK8sAPI) DeletePodsWithRetry(clientset kubernetes.Interface, namespace, podNameContains string) error {
	return nil
}
func (m *MockK8sAPI) DeletePods(clientset kubernetes.Interface, namespace, podNameContains string) error {
	return nil
}
func (m *MockK8sAPI) GetPodLogs(clientset kubernetes.Interface, podName, namespace, container string, fromTime time.Time) (string, error) {
	return "", nil
}
func (m *MockK8sAPI) RestartContainers(clientset kubernetes.Interface, kubeconfigpath, namespace, podnamecontains, containerName string) (string, error) {
	return "", nil
}
func (m *MockK8sAPI) GetLogFromFirstPod(clientset kubernetes.Interface, namespace, podNameContains, containerName string, logsFromTime time.Time) (string, error) {
	return "", nil
}

// Network policy methods
func (m *MockK8sAPI) CreateNetworkPolicy(clientset kubernetes.Interface, networkpolicy *networkingv1.NetworkPolicy, namespace string) error {
	return nil
}

// Namespace methods
func (m *MockK8sAPI) GetNamespaces(clientset kubernetes.Interface) (*corev1.NamespaceList, error) {
	return nil, nil
}
func (m *MockK8sAPI) GetNamespaceWithLabelFilter(clientset kubernetes.Interface, labels map[string]string) ([]string, error) {
	return nil, nil
}

// Job methods
func (m *MockK8sAPI) CreateJob(clientset kubernetes.Interface, job *batchv1.Job, namespace string) error {
	return nil
}
func (m *MockK8sAPI) DeleteJob(clientset kubernetes.Interface, namespace, jobName string) error {
	return nil
}
func (m *MockK8sAPI) DeleteJobWithPods(clientset kubernetes.Interface, namespace, jobName string) error {
	return nil
}

// Deployment methods
func (m *MockK8sAPI) DeleteDeployment(clientset kubernetes.Interface, namespace, deploymentName string) error {
	return nil
}
func (m *MockK8sAPI) GetDeployments(clientset kubernetes.Interface, namespace string) (*appsv1.DeploymentList, error) {
	return nil, nil
}
func (m *MockK8sAPI) GetDeployment(clientset kubernetes.Interface, namespace, deploymentName string) (appsv1.Deployment, error) {
	return appsv1.Deployment{}, nil
}
func (m *MockK8sAPI) GetDeploymentWithSpecAnnotationFilter(clientset kubernetes.Interface, namespace string, annotationName map[string]string) ([]string, error) {
	return nil, nil
}

// Argo Rollout methods
func (m *MockK8sAPI) GetArgoRollouts(clientset argo.Interface, namespace string) (*argov1alpha1.RolloutList, error) {
	return nil, nil
}
func (m *MockK8sAPI) GetArgoRolloutsWithSpecAnnotationFilter(clientset argo.Interface, namespace string, annotations map[string]string) ([]string, error) {
	return nil, nil
}

// ConfigMap methods
func (m *MockK8sAPI) ReadConfigMap(clientset kubernetes.Interface, namespace, configMapName string) (*corev1.ConfigMap, error) {
	return nil, nil
}
func (m *MockK8sAPI) ReadConfigMapData(clientset kubernetes.Interface, namespace, configMapName string) (map[string]string, error) {
	return nil, nil
}

// TestRestartDeploymentWrapper_RestartObject tests core restart functionality
func TestRestartDeploymentWrapper_RestartObject(t *testing.T) {
	tests := []struct {
		name              string
		restartConfig     crdv1.StatusCheckConfig
		expectStatusCheck bool
	}{
		{
			name: "restart without status check",
			restartConfig: crdv1.StatusCheckConfig{
				CheckStatus: false,
			},
			expectStatusCheck: false,
		},
		{
			name: "restart with status check enabled",
			restartConfig: crdv1.StatusCheckConfig{
				CheckStatus:   true,
				MaxRetry:      3,
				RetryDuration: 1,
			},
			expectStatusCheck: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockK8s := &MockK8sAPI{
				RolloutRestartDeploymentFunc: func(kubeconfig, ns, name string) (string, error) {
					return "deployment.apps/" + name + " restarted", nil
				},
				RolloutDeploymentStatusFunc: func(kubeconfig, ns, name string) (string, error) {
					return "deployment \"" + name + "\" successfully rolled out", nil
				},
			}

			wrapper := RestartDeploymentWrapper{}
			wrapper.RestartObject(mockK8s, tt.restartConfig, "test-ns", "test-deployment", 0)

			// Verify restart was called
			if len(mockK8s.RestartDeploymentCalls) != 1 {
				t.Errorf("Expected 1 restart call, got %d", len(mockK8s.RestartDeploymentCalls))
			}

			// Verify status check behavior
			if tt.expectStatusCheck {
				if len(mockK8s.DeploymentStatusCalls) == 0 {
					t.Error("Expected status check calls but got none")
				}
			} else {
				if len(mockK8s.DeploymentStatusCalls) > 0 {
					t.Errorf("Expected no status check calls but got %d", len(mockK8s.DeploymentStatusCalls))
				}
			}
		})
	}
}

// Test the actual Restart methods with real function calls
func TestRestartDeploymentWrapper_Restart_ActualCalls(t *testing.T) {
	tests := []struct {
		name             string
		restartType      string
		namespaceObjects map[string][]string
		expectCalls      int
	}{
		{
			name:        "valid deployment type",
			restartType: common.DEPLOYMENT,
			namespaceObjects: map[string][]string{
				"test-ns": {"deployment1", "deployment2"},
			},
			expectCalls: 2, // 2 deployments
		},
		{
			name:        "multiple namespaces",
			restartType: common.DEPLOYMENT,
			namespaceObjects: map[string][]string{
				"ns1": {"dep1"},
				"ns2": {"dep2", "dep3"},
			},
			expectCalls: 3, // 3 total deployments
		},
		{
			name:             "empty namespace objects",
			restartType:      common.DEPLOYMENT,
			namespaceObjects: map[string][]string{},
			expectCalls:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockK8s := &MockK8sAPI{
				RolloutRestartDeploymentFunc: func(kubeconfig, ns, name string) (string, error) {
					return "deployment.apps/" + name + " restarted", nil
				},
			}

			restartObjects := common.RestartObjects{
				Type:             tt.restartType,
				NamespaceObjects: tt.namespaceObjects,
				RestartConfig: crdv1.StatusCheckConfig{
					CheckStatus:   false,
					MaxRetry:      0,
					RetryDuration: 0,
				},
			}

			wrapper := RestartDeploymentWrapper{}
			wrapper.Restart(mockK8s, restartObjects)

			// Verify the correct number of restart calls were made
			if len(mockK8s.RestartDeploymentCalls) != tt.expectCalls {
				t.Errorf("Expected %d restart calls, got %d", tt.expectCalls, len(mockK8s.RestartDeploymentCalls))
			}

			// Verify all expected objects were restarted
			callMap := make(map[string][]string)
			for _, call := range mockK8s.RestartDeploymentCalls {
				callMap[call.Namespace] = append(callMap[call.Namespace], call.ObjectName)
			}

			for expectedNs, expectedObjects := range tt.namespaceObjects {
				actualObjects, exists := callMap[expectedNs]
				if !exists && len(expectedObjects) > 0 {
					t.Errorf("Expected calls for namespace %s but got none", expectedNs)
					continue
				}
				if len(actualObjects) != len(expectedObjects) {
					t.Errorf("Expected %d objects in namespace %s, got %d", len(expectedObjects), expectedNs, len(actualObjects))
				}
			}
		})
	}
}

// Test RestartRolloutWrapper methods
func TestRestartRolloutWrapper_Restart_ActualCalls(t *testing.T) {
	tests := []struct {
		name             string
		restartType      string
		namespaceObjects map[string][]string
		expectCalls      int
	}{
		{
			name:        "valid argorollout type",
			restartType: common.ARGO_ROLLOUT,
			namespaceObjects: map[string][]string{
				"test-ns": {"rollout1", "rollout2"},
			},
			expectCalls: 2,
		},
		{
			name:        "multiple namespaces",
			restartType: common.ARGO_ROLLOUT,
			namespaceObjects: map[string][]string{
				"ns1": {"rollout1"},
				"ns2": {"rollout2", "rollout3"},
			},
			expectCalls: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockK8s := &MockK8sAPI{
				RolloutRestartArgoRolloutsFunc: func(kubeconfig, ns, name string) (string, error) {
					return "rollout.argoproj.io/" + name + " restarted", nil
				},
			}

			restartObjects := common.RestartObjects{
				Type:             tt.restartType,
				NamespaceObjects: tt.namespaceObjects,
				RestartConfig: crdv1.StatusCheckConfig{
					CheckStatus:   false,
					MaxRetry:      0,
					RetryDuration: 0,
				},
			}

			wrapper := RestartRolloutWrapper{}
			wrapper.Restart(mockK8s, restartObjects)

			// Verify the correct number of restart calls were made
			if len(mockK8s.RestartRolloutCalls) != tt.expectCalls {
				t.Errorf("Expected %d restart calls, got %d", tt.expectCalls, len(mockK8s.RestartRolloutCalls))
			}
		})
	}
}

// TestRestartRolloutWrapper_RestartObject tests core rollout restart functionality
func TestRestartRolloutWrapper_RestartObject(t *testing.T) {
	mockK8s := &MockK8sAPI{
		RolloutRestartArgoRolloutsFunc: func(kubeconfig, ns, name string) (string, error) {
			return "rollout.argoproj.io/" + name + " restarted", nil
		},
		RolloutArogRolloutStatusFunc: func(kubeconfig, ns, name string) (string, error) {
			return "Healthy", nil
		},
	}

	// Test with status check enabled (most complex scenario)
	config := crdv1.StatusCheckConfig{
		CheckStatus:   true,
		MaxRetry:      3,
		RetryDuration: 1,
	}

	wrapper := RestartRolloutWrapper{}
	wrapper.RestartObject(mockK8s, config, "test-ns", "test-rollout", 0)

	// Verify restart and status check were called
	if len(mockK8s.RestartRolloutCalls) != 1 {
		t.Errorf("Expected 1 restart call, got %d", len(mockK8s.RestartRolloutCalls))
	}
	if len(mockK8s.RolloutStatusCalls) == 0 {
		t.Error("Expected status check calls but got none")
	}
}

// TestRestartDeploymentWrapper_WaitForRestartToBeComplete tests critical retry scenarios
func TestRestartDeploymentWrapper_WaitForRestartToBeComplete(t *testing.T) {
	tests := []struct {
		name          string
		maxRetry      int
		statusOutputs []string
		expectRetries int
	}{
		{
			name:          "success after retries (production risk scenario)",
			maxRetry:      3,
			statusOutputs: []string{"waiting", "still waiting", "successfully rolled out"},
			expectRetries: 3,
		},
		{
			name:          "timeout after max retries (critical failure scenario)",
			maxRetry:      2,
			statusOutputs: []string{"waiting", "still waiting", "still waiting"},
			expectRetries: 3, // Initial + 2 retries
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			mockK8s := &MockK8sAPI{
				RolloutDeploymentStatusFunc: func(kubeconfig, ns, name string) (string, error) {
					if callCount < len(tt.statusOutputs) {
						output := tt.statusOutputs[callCount]
						callCount++
						return output, nil
					}
					return "still waiting", nil
				},
			}

			config := crdv1.StatusCheckConfig{
				CheckStatus:   true,
				MaxRetry:      tt.maxRetry,
				RetryDuration: 1,
			}

			wrapper := RestartDeploymentWrapper{}
			wrapper.WaitForRestartToBeComplete(mockK8s, config, "test-ns", "test-deployment", 0)

			// Verify retry logic worked correctly
			if len(mockK8s.DeploymentStatusCalls) != tt.expectRetries {
				t.Errorf("Expected %d status calls, got %d", tt.expectRetries, len(mockK8s.DeploymentStatusCalls))
			}
		})
	}
}

// TestRestartRolloutWrapper_WaitForRestartToBeComplete tests rollout retry logic
func TestRestartRolloutWrapper_WaitForRestartToBeComplete(t *testing.T) {
	// Test the most critical scenario: success after retries
	callCount := 0
	statusOutputs := []string{"Progressing", "Progressing", "Healthy"}

	mockK8s := &MockK8sAPI{
		RolloutArogRolloutStatusFunc: func(kubeconfig, ns, name string) (string, error) {
			if callCount < len(statusOutputs) {
				output := statusOutputs[callCount]
				callCount++
				return output, nil
			}
			return "Progressing", nil
		},
	}

	config := crdv1.StatusCheckConfig{
		CheckStatus:   true,
		MaxRetry:      3,
		RetryDuration: 1,
	}

	wrapper := RestartRolloutWrapper{}
	wrapper.WaitForRestartToBeComplete(mockK8s, config, "test-ns", "test-rollout", 0)

	// Verify retry logic worked correctly
	if len(mockK8s.RolloutStatusCalls) != 3 {
		t.Errorf("Expected 3 status calls, got %d", len(mockK8s.RolloutStatusCalls))
	}
}

// withStagger toggles common.StaggerNamespacesEnabled for the duration of a test.
func withStagger(t *testing.T, enabled bool) {
	t.Helper()
	original := common.StaggerNamespacesEnabled
	common.StaggerNamespacesEnabled = enabled
	t.Cleanup(func() { common.StaggerNamespacesEnabled = original })
}

// staggerTestObjects deliberately uses keys that sort differently from insertion order.
var staggerTestObjects = map[string][]string{
	"ns-c": {"obj-c1"},
	"ns-a": {"obj-a1", "obj-a2"},
	"ns-b": {"obj-b1"},
}

// assertStaggeredSequence checks that, for sorted namespaces, every restart and
// every status call of namespace n precedes any call for namespace n+1, and that
// each object got exactly one restart and at least one status check.
func assertStaggeredSequence(t *testing.T, sequence []string, objects map[string][]string) {
	t.Helper()
	wantOrder := []string{"ns-a", "ns-b", "ns-c"}

	// Namespace of each call, de-duplicated consecutively, must equal the sorted order.
	var seenOrder []string
	for _, call := range sequence {
		ns := strings.SplitN(strings.SplitN(call, ":", 2)[1], "/", 2)[0]
		if len(seenOrder) == 0 || seenOrder[len(seenOrder)-1] != ns {
			seenOrder = append(seenOrder, ns)
		}
	}
	if !reflect.DeepEqual(seenOrder, wantOrder) {
		t.Fatalf("namespaces were not processed in sorted, gated order.\nwant %v\ngot  %v\nsequence: %v", wantOrder, seenOrder, sequence)
	}

	// Within a namespace: all restarts, then at least one status per object.
	for ns, names := range objects {
		restarts, statuses := 0, map[string]int{}
		lastRestartIdx, firstStatusIdx := -1, len(sequence)
		for i, call := range sequence {
			if !strings.Contains(call, ":"+ns+"/") {
				continue
			}
			if strings.HasPrefix(call, "restart:") {
				restarts++
				lastRestartIdx = i
			} else {
				statuses[strings.TrimPrefix(call, "status:"+ns+"/")]++
				if i < firstStatusIdx {
					firstStatusIdx = i
				}
			}
		}
		if restarts != len(names) {
			t.Errorf("namespace %s: expected %d restarts, got %d", ns, len(names), restarts)
		}
		if lastRestartIdx > firstStatusIdx {
			t.Errorf("namespace %s: a restart happened after a status check began (sequence %v)", ns, sequence)
		}
		for _, name := range names {
			if statuses[name] == 0 {
				t.Errorf("namespace %s: object %s was never status-checked", ns, name)
			}
		}
	}
}

func TestRestartDeploymentWrapper_Restart_StaggerOn_OrdersAndGatesEvenWithCheckStatusFalse(t *testing.T) {
	withStagger(t, true)
	mockK8s := &MockK8sAPI{}

	RestartDeploymentWrapper{}.Restart(mockK8s, common.RestartObjects{
		Type:             common.DEPLOYMENT,
		NamespaceObjects: staggerTestObjects,
		// CheckStatus false must NOT disable the per-namespace gate.
		RestartConfig: crdv1.StatusCheckConfig{CheckStatus: false, MaxRetry: 2, RetryDuration: 1},
	})

	if len(mockK8s.RestartDeploymentCalls) != 4 {
		t.Fatalf("expected 4 restarts, got %d", len(mockK8s.RestartDeploymentCalls))
	}
	assertStaggeredSequence(t, mockK8s.Sequence, staggerTestObjects)
}

func TestRestartRolloutWrapper_Restart_StaggerOn_OrdersAndGatesEvenWithCheckStatusFalse(t *testing.T) {
	withStagger(t, true)
	mockK8s := &MockK8sAPI{}

	RestartRolloutWrapper{}.Restart(mockK8s, common.RestartObjects{
		Type:             common.ARGO_ROLLOUT,
		NamespaceObjects: staggerTestObjects,
		RestartConfig:    crdv1.StatusCheckConfig{CheckStatus: false, MaxRetry: 2, RetryDuration: 1},
	})

	if len(mockK8s.RestartRolloutCalls) != 4 {
		t.Fatalf("expected 4 restarts, got %d", len(mockK8s.RestartRolloutCalls))
	}
	assertStaggeredSequence(t, mockK8s.Sequence, staggerTestObjects)
}

// With CheckStatus already true, staggering must not double-wait: each object is
// status-checked once (healthy on first poll), not once in RestartObject and again in the gate.
func TestRestartDeploymentWrapper_Restart_StaggerOn_NoDoubleWaitWhenCheckStatusTrue(t *testing.T) {
	withStagger(t, true)
	mockK8s := &MockK8sAPI{}

	RestartDeploymentWrapper{}.Restart(mockK8s, common.RestartObjects{
		Type:             common.DEPLOYMENT,
		NamespaceObjects: map[string][]string{"ns-a": {"obj-a1", "obj-a2"}},
		RestartConfig:    crdv1.StatusCheckConfig{CheckStatus: true, MaxRetry: 2, RetryDuration: 1},
	})

	if len(mockK8s.DeploymentStatusCalls) != 2 {
		t.Errorf("expected exactly 2 status calls (one per object), got %d: %v", len(mockK8s.DeploymentStatusCalls), mockK8s.Sequence)
	}
}

// The gate must hold the next namespace until the current one is healthy or retries are exhausted.
func TestRestartDeploymentWrapper_Restart_StaggerOn_WaitsForUnhealthyNamespaceUntilRetriesExhausted(t *testing.T) {
	withStagger(t, true)
	mockK8s := &MockK8sAPI{
		RolloutDeploymentStatusFunc: func(_, ns, name string) (string, error) {
			if ns == "ns-a" {
				return "Waiting for deployment spec update to be observed...", nil
			}
			return "successfully rolled out", nil
		},
	}

	RestartDeploymentWrapper{}.Restart(mockK8s, common.RestartObjects{
		Type:             common.DEPLOYMENT,
		NamespaceObjects: map[string][]string{"ns-a": {"obj-a1"}, "ns-b": {"obj-b1"}},
		RestartConfig:    crdv1.StatusCheckConfig{CheckStatus: false, MaxRetry: 2, RetryDuration: 1},
	})

	// ns-a: 1 restart + 3 status polls (initial + 2 retries); ns-b: 1 restart + 1 status.
	want := []string{
		"restart:ns-a/obj-a1", "status:ns-a/obj-a1", "status:ns-a/obj-a1", "status:ns-a/obj-a1",
		"restart:ns-b/obj-b1", "status:ns-b/obj-b1",
	}
	if !reflect.DeepEqual(mockK8s.Sequence, want) {
		t.Errorf("unexpected sequence\nwant %v\ngot  %v", want, mockK8s.Sequence)
	}
}

func TestRestart_StaggerOff_BehaviorUnchanged(t *testing.T) {
	withStagger(t, false)

	t.Run("deployment: no status calls when CheckStatus false", func(t *testing.T) {
		mockK8s := &MockK8sAPI{}
		RestartDeploymentWrapper{}.Restart(mockK8s, common.RestartObjects{
			Type:             common.DEPLOYMENT,
			NamespaceObjects: staggerTestObjects,
			RestartConfig:    crdv1.StatusCheckConfig{CheckStatus: false},
		})
		if len(mockK8s.RestartDeploymentCalls) != 4 {
			t.Errorf("expected 4 restarts, got %d", len(mockK8s.RestartDeploymentCalls))
		}
		if len(mockK8s.DeploymentStatusCalls) != 0 {
			t.Errorf("expected no status calls with staggering off, got %d", len(mockK8s.DeploymentStatusCalls))
		}
	})

	t.Run("rollout: no status calls when CheckStatus false", func(t *testing.T) {
		mockK8s := &MockK8sAPI{}
		RestartRolloutWrapper{}.Restart(mockK8s, common.RestartObjects{
			Type:             common.ARGO_ROLLOUT,
			NamespaceObjects: staggerTestObjects,
			RestartConfig:    crdv1.StatusCheckConfig{CheckStatus: false},
		})
		if len(mockK8s.RestartRolloutCalls) != 4 {
			t.Errorf("expected 4 restarts, got %d", len(mockK8s.RestartRolloutCalls))
		}
		if len(mockK8s.RolloutStatusCalls) != 0 {
			t.Errorf("expected no status calls with staggering off, got %d", len(mockK8s.RolloutStatusCalls))
		}
	})
}

func TestStaggerWaitConfig(t *testing.T) {
	tests := []struct {
		name string
		in   crdv1.StatusCheckConfig
		want crdv1.StatusCheckConfig
	}{
		{
			name: "defaults when unset",
			in:   crdv1.StatusCheckConfig{},
			want: crdv1.StatusCheckConfig{CheckStatus: true, MaxRetry: common.DefaultStaggerMaxRetry, RetryDuration: common.DefaultStaggerRetryDurationSeconds},
		},
		{
			name: "uses CRD values and forces CheckStatus",
			in:   crdv1.StatusCheckConfig{CheckStatus: false, MaxRetry: 4, RetryDuration: 7},
			want: crdv1.StatusCheckConfig{CheckStatus: true, MaxRetry: 4, RetryDuration: 7},
		},
		{
			name: "partial: only MaxRetry set",
			in:   crdv1.StatusCheckConfig{MaxRetry: 3},
			want: crdv1.StatusCheckConfig{CheckStatus: true, MaxRetry: 3, RetryDuration: common.DefaultStaggerRetryDurationSeconds},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := staggerWaitConfig(tt.in); got != tt.want {
				t.Errorf("staggerWaitConfig(%+v) = %+v, want %+v", tt.in, got, tt.want)
			}
		})
	}
}

// Regression: rollout wait previously recursed without incrementing retryCount and never timed out.
func TestRestartRolloutWrapper_WaitForRestartToBeComplete_TimesOutAfterMaxRetry(t *testing.T) {
	mockK8s := &MockK8sAPI{
		RolloutArogRolloutStatusFunc: func(_, _, _ string) (string, error) { return "Progressing", nil },
	}
	RestartRolloutWrapper{}.WaitForRestartToBeComplete(mockK8s, crdv1.StatusCheckConfig{CheckStatus: true, MaxRetry: 2, RetryDuration: 1}, "test-ns", "test-rollout", 0)

	if len(mockK8s.RolloutStatusCalls) != 3 { // initial + 2 retries
		t.Errorf("expected 3 status calls, got %d", len(mockK8s.RolloutStatusCalls))
	}
}
