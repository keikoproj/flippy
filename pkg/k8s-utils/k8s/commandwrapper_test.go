package k8s

import (
	"testing"
	"time"

	argov1alpha1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	argo "github.com/argoproj/argo-rollouts/pkg/client/clientset/versioned"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// MockK8sAPI for testing kubectl command functions
type MockK8sAPI struct {
	mock.Mock
}

func (m *MockK8sAPI) ExecuteKubectlCommand(cmdParameter []string) (string, error) {
	args := m.Called(cmdParameter)
	return args.String(0), args.Error(1)
}

func (m *MockK8sAPI) PatchResource(kubeconfigpath string, namespace string, resource string, resourceName string, patchJson string) (string, error) {
	args := m.Called(kubeconfigpath, namespace, resource, resourceName, patchJson)
	return args.String(0), args.Error(1)
}

func (m *MockK8sAPI) ScaleDeployment(kubeconfigpath string, namespace string, deploymentname string, scale int) (string, error) {
	args := m.Called(kubeconfigpath, namespace, deploymentname, scale)
	return args.String(0), args.Error(1)
}

func (m *MockK8sAPI) ApplyYaml(kubeconfigpath string, yamlFilePath string) (string, error) {
	args := m.Called(kubeconfigpath, yamlFilePath)
	return args.String(0), args.Error(1)
}

func (m *MockK8sAPI) DeleteYaml(kubeconfigpath string, yamlFilePath string) (string, error) {
	args := m.Called(kubeconfigpath, yamlFilePath)
	return args.String(0), args.Error(1)
}

func (m *MockK8sAPI) GetServiceEntries(kubeconfigpath string, namespace string) (map[string]string, error) {
	args := m.Called(kubeconfigpath, namespace)
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockK8sAPI) RestartContainer(kubeconfigpath string, namespace string, podname string, containerName string) (string, error) {
	args := m.Called(kubeconfigpath, namespace, podname, containerName)
	return args.String(0), args.Error(1)
}

func (m *MockK8sAPI) RolloutRestartDeployment(kubeconfigpath string, namespace string, deploymentName string) (string, error) {
	args := m.Called(kubeconfigpath, namespace, deploymentName)
	return args.String(0), args.Error(1)
}

func (m *MockK8sAPI) RolloutDeploymentStatus(kubeconfigpath string, namespace string, deploymentName string) (string, error) {
	args := m.Called(kubeconfigpath, namespace, deploymentName)
	return args.String(0), args.Error(1)
}

// Pod-related methods
func (m *MockK8sAPI) GetRunningPods(clientset kubernetes.Interface, namespace string, podNameContains string) ([]corev1.Pod, error) {
	args := m.Called(clientset, namespace, podNameContains)
	return args.Get(0).([]corev1.Pod), args.Error(1)
}

func (m *MockK8sAPI) GetAllPodsInNamespace(clientset kubernetes.Interface, namespace string) ([]corev1.Pod, error) {
	args := m.Called(clientset, namespace)
	return args.Get(0).([]corev1.Pod), args.Error(1)
}

func (m *MockK8sAPI) DeletePod(clientset kubernetes.Interface, namespace string, podName string) error {
	args := m.Called(clientset, namespace, podName)
	return args.Error(0)
}

func (m *MockK8sAPI) DeletePodsWithRetry(clientset kubernetes.Interface, namespace string, podNameContains string) error {
	args := m.Called(clientset, namespace, podNameContains)
	return args.Error(0)
}

func (m *MockK8sAPI) DeletePods(clientset kubernetes.Interface, namespace string, podNameContains string) error {
	args := m.Called(clientset, namespace, podNameContains)
	return args.Error(0)
}

func (m *MockK8sAPI) GetPodLogs(clientset kubernetes.Interface, podName string, namespace string, container string, fromTime time.Time) (string, error) {
	args := m.Called(clientset, podName, namespace, container, fromTime)
	return args.String(0), args.Error(1)
}

func (m *MockK8sAPI) RestartContainers(clientset kubernetes.Interface, kubeconfigpath string, namespace string, podnamecontains string, containerName string) (string, error) {
	args := m.Called(clientset, kubeconfigpath, namespace, podnamecontains, containerName)
	return args.String(0), args.Error(1)
}

func (m *MockK8sAPI) GetLogFromFirstPod(clientset kubernetes.Interface, namespace string, podNameContains string, containerName string, logsFromTime time.Time) (string, error) {
	args := m.Called(clientset, namespace, podNameContains, containerName, logsFromTime)
	return args.String(0), args.Error(1)
}

// NetworkPolicy methods
func (m *MockK8sAPI) CreateNetworkPolicy(clientset kubernetes.Interface, networkpolicy *networkingv1.NetworkPolicy, namespace string) error {
	args := m.Called(clientset, networkpolicy, namespace)
	return args.Error(0)
}

// Namespace methods
func (m *MockK8sAPI) GetNamespaces(clientset kubernetes.Interface) (*corev1.NamespaceList, error) {
	args := m.Called(clientset)
	return args.Get(0).(*corev1.NamespaceList), args.Error(1)
}

func (m *MockK8sAPI) GetNamespaceWithLabelFilter(clientset kubernetes.Interface, labels map[string]string) ([]string, error) {
	args := m.Called(clientset, labels)
	return args.Get(0).([]string), args.Error(1)
}

// Job methods
func (m *MockK8sAPI) CreateJob(clientset kubernetes.Interface, job *batchv1.Job, namespace string) error {
	args := m.Called(clientset, job, namespace)
	return args.Error(0)
}

func (m *MockK8sAPI) DeleteJob(clientset kubernetes.Interface, namespace string, jobName string) error {
	args := m.Called(clientset, namespace, jobName)
	return args.Error(0)
}

func (m *MockK8sAPI) DeleteJobWithPods(clientset kubernetes.Interface, namespace string, jobName string) error {
	args := m.Called(clientset, namespace, jobName)
	return args.Error(0)
}

// Deployment methods
func (m *MockK8sAPI) DeleteDeployment(clientset kubernetes.Interface, namespace string, deploymentName string) error {
	args := m.Called(clientset, namespace, deploymentName)
	return args.Error(0)
}

func (m *MockK8sAPI) GetDeployments(clientset kubernetes.Interface, namespace string) (*appsv1.DeploymentList, error) {
	args := m.Called(clientset, namespace)
	return args.Get(0).(*appsv1.DeploymentList), args.Error(1)
}

func (m *MockK8sAPI) GetDeployment(clientset kubernetes.Interface, namespace string, deploymentName string) (appsv1.Deployment, error) {
	args := m.Called(clientset, namespace, deploymentName)
	return args.Get(0).(appsv1.Deployment), args.Error(1)
}

func (m *MockK8sAPI) GetDeploymentWithSpecAnnotationFilter(clientset kubernetes.Interface, namespace string, annotationName map[string]string) ([]string, error) {
	args := m.Called(clientset, namespace, annotationName)
	return args.Get(0).([]string), args.Error(1)
}

// ArgoRollouts methods
func (m *MockK8sAPI) GetArgoRollouts(clientset argo.Interface, namespace string) (*argov1alpha1.RolloutList, error) {
	args := m.Called(clientset, namespace)
	return args.Get(0).(*argov1alpha1.RolloutList), args.Error(1)
}

func (m *MockK8sAPI) GetArgoRolloutsWithSpecAnnotationFilter(clientset argo.Interface, namespace string, annotations map[string]string) ([]string, error) {
	args := m.Called(clientset, namespace, annotations)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockK8sAPI) RolloutRestartArgoRollouts(kubeconfigpath string, namespace string, argoRolloutName string) (string, error) {
	args := m.Called(kubeconfigpath, namespace, argoRolloutName)
	return args.String(0), args.Error(1)
}

func (m *MockK8sAPI) RolloutArogRolloutStatus(kubeconfigpath string, namespace string, argoRolloutName string) (string, error) {
	args := m.Called(kubeconfigpath, namespace, argoRolloutName)
	return args.String(0), args.Error(1)
}

// ConfigMap methods
func (m *MockK8sAPI) ReadConfigMap(clientset kubernetes.Interface, namespace string, configMapName string) (*corev1.ConfigMap, error) {
	args := m.Called(clientset, namespace, configMapName)
	return args.Get(0).(*corev1.ConfigMap), args.Error(1)
}

func (m *MockK8sAPI) ReadConfigMapData(clientset kubernetes.Interface, namespace string, configMapName string) (map[string]string, error) {
	args := m.Called(clientset, namespace, configMapName)
	return args.Get(0).(map[string]string), args.Error(1)
}

func TestExecuteKubectlCommand_Success(t *testing.T) {
	// Test successful kubectl command execution
	wrapper := K8sWrapper{}

	// Test with simple command that should work in most environments
	cmdParams := []string{"version", "--client"}

	output, err := wrapper.ExecuteKubectlCommand(cmdParams)

	// We expect this to either succeed or fail with a specific kubectl error
	// The important thing is that our error handling works correctly
	if err != nil {
		// If kubectl is not available, we should get a proper error message
		assert.Contains(t, err.Error(), "Failed to execute - kubectl")
		assert.Contains(t, err.Error(), "version --client")
	} else {
		// If kubectl is available, we should get some output
		assert.NotEmpty(t, output)
	}
}

func TestExecuteKubectlCommand_ErrorHandling(t *testing.T) {
	// Test error handling with invalid command
	wrapper := K8sWrapper{}

	cmdParams := []string{"invalid-command", "--nonexistent-flag"}

	output, err := wrapper.ExecuteKubectlCommand(cmdParams)

	// Should return an error with proper formatting
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to execute - kubectl")
	assert.Contains(t, err.Error(), "invalid-command --nonexistent-flag")
	assert.NotEmpty(t, output) // Should contain error output from kubectl
}

func TestPatchResource(t *testing.T) {
	// Create mock
	mockAPI := &MockK8sAPI{}
	originalK8s := K8s
	K8s = mockAPI
	defer func() { K8s = originalK8s }()

	// Setup expectations
	expectedArgs := []string{"patch", "--kubeconfig=/path/to/config", "-n", "test-namespace", "deployment", "test-deployment", "-p", `{"spec":{"replicas":3}}`, "--type", "json"}
	mockAPI.On("ExecuteKubectlCommand", expectedArgs).Return("deployment.apps/test-deployment patched", nil)

	wrapper := K8sWrapper{}
	output, err := wrapper.PatchResource("/path/to/config", "test-namespace", "deployment", "test-deployment", `{"spec":{"replicas":3}}`)

	assert.NoError(t, err)
	assert.Equal(t, "deployment.apps/test-deployment patched", output)
	mockAPI.AssertExpectations(t)
}

func TestScaleDeployment(t *testing.T) {
	// Create mock
	mockAPI := &MockK8sAPI{}
	originalK8s := K8s
	K8s = mockAPI
	defer func() { K8s = originalK8s }()

	// Setup expectations
	expectedArgs := []string{"scale", "--kubeconfig=/path/to/config", "-n", "test-namespace", "deployment/test-deployment", "--replicas=5"}
	mockAPI.On("ExecuteKubectlCommand", expectedArgs).Return("deployment.apps/test-deployment scaled", nil)

	wrapper := K8sWrapper{}
	output, err := wrapper.ScaleDeployment("/path/to/config", "test-namespace", "test-deployment", 5)

	assert.NoError(t, err)
	assert.Equal(t, "deployment.apps/test-deployment scaled", output)
	mockAPI.AssertExpectations(t)
}

func TestApplyYaml(t *testing.T) {
	// Create mock
	mockAPI := &MockK8sAPI{}
	originalK8s := K8s
	K8s = mockAPI
	defer func() { K8s = originalK8s }()

	// Setup expectations
	expectedArgs := []string{"--kubeconfig=/path/to/config", "apply", "-f", "/path/to/manifest.yaml"}
	mockAPI.On("ExecuteKubectlCommand", expectedArgs).Return("deployment.apps/test-deployment created", nil)

	wrapper := K8sWrapper{}
	output, err := wrapper.ApplyYaml("/path/to/config", "/path/to/manifest.yaml")

	assert.NoError(t, err)
	assert.Equal(t, "deployment.apps/test-deployment created", output)
	mockAPI.AssertExpectations(t)
}

func TestDeleteYaml(t *testing.T) {
	// Create mock
	mockAPI := &MockK8sAPI{}
	originalK8s := K8s
	K8s = mockAPI
	defer func() { K8s = originalK8s }()

	// Setup expectations
	expectedArgs := []string{"--kubeconfig=/path/to/config", "delete", "-f", "/path/to/manifest.yaml"}
	mockAPI.On("ExecuteKubectlCommand", expectedArgs).Return("deployment.apps/test-deployment deleted", nil)

	wrapper := K8sWrapper{}
	output, err := wrapper.DeleteYaml("/path/to/config", "/path/to/manifest.yaml")

	assert.NoError(t, err)
	assert.Equal(t, "deployment.apps/test-deployment deleted", output)
	mockAPI.AssertExpectations(t)
}

func TestRolloutRestartDeployment(t *testing.T) {
	// Create mock
	mockAPI := &MockK8sAPI{}
	originalK8s := K8s
	K8s = mockAPI
	defer func() { K8s = originalK8s }()

	// Setup expectations
	expectedArgs := []string{"--kubeconfig=/path/to/config", "-n", "test-namespace", "rollout", "restart", "deployment", "test-deployment"}
	mockAPI.On("ExecuteKubectlCommand", expectedArgs).Return("deployment.apps/test-deployment restarted", nil)

	wrapper := K8sWrapper{}
	output, err := wrapper.RolloutRestartDeployment("/path/to/config", "test-namespace", "test-deployment")

	assert.NoError(t, err)
	assert.Equal(t, "deployment.apps/test-deployment restarted", output)
	mockAPI.AssertExpectations(t)
}

func TestRolloutDeploymentStatus(t *testing.T) {
	// Create mock
	mockAPI := &MockK8sAPI{}
	originalK8s := K8s
	K8s = mockAPI
	defer func() { K8s = originalK8s }()

	// Setup expectations
	expectedArgs := []string{"--kubeconfig=/path/to/config", "-n", "test-namespace", "rollout", "status", "deployment", "test-deployment", "--watch=false"}
	mockAPI.On("ExecuteKubectlCommand", expectedArgs).Return("deployment \"test-deployment\" successfully rolled out", nil)

	wrapper := K8sWrapper{}
	output, err := wrapper.RolloutDeploymentStatus("/path/to/config", "test-namespace", "test-deployment")

	assert.NoError(t, err)
	assert.Equal(t, "deployment \"test-deployment\" successfully rolled out", output)
	mockAPI.AssertExpectations(t)
}

func TestRestartContainer(t *testing.T) {
	// Create mock
	mockAPI := &MockK8sAPI{}
	originalK8s := K8s
	K8s = mockAPI
	defer func() { K8s = originalK8s }()

	// Setup expectations
	expectedArgs := []string{"exec", "-it", "test-pod", "--kubeconfig=/path/to/config", "-n", "test-namespace", "-c", "test-container", "--", "/bin/bash", "-c", "kill 1"}
	mockAPI.On("ExecuteKubectlCommand", expectedArgs).Return("", nil)

	wrapper := K8sWrapper{}
	output, err := wrapper.RestartContainer("/path/to/config", "test-namespace", "test-pod", "test-container")

	assert.NoError(t, err)
	assert.Equal(t, "", output)
	mockAPI.AssertExpectations(t)
}

func TestGetServiceEntries_Success(t *testing.T) {
	// Create mock
	mockAPI := &MockK8sAPI{}
	originalK8s := K8s
	K8s = mockAPI
	defer func() { K8s = originalK8s }()

	// Mock kubectl output for service entries
	kubectlOutput := `NAME                    HOSTS                   LOCATION   RESOLUTION   AGE
test-service-entry      [test.example.com]      MESH_EXTERNAL   DNS          1d
another-service         [api.example.com]       MESH_EXTERNAL   DNS          2d`

	expectedArgs := []string{"--kubeconfig=/path/to/config", "get", "se", "-n", "test-namespace"}
	mockAPI.On("ExecuteKubectlCommand", expectedArgs).Return(kubectlOutput, nil)

	wrapper := K8sWrapper{}
	serviceEntries, err := wrapper.GetServiceEntries("/path/to/config", "test-namespace")

	assert.NoError(t, err)
	assert.Len(t, serviceEntries, 2)
	assert.Equal(t, "test.example.com", serviceEntries["test-service-entry"])
	assert.Equal(t, "api.example.com", serviceEntries["another-service"])
	mockAPI.AssertExpectations(t)
}

func TestGetServiceEntries_Error(t *testing.T) {
	// Create mock
	mockAPI := &MockK8sAPI{}
	originalK8s := K8s
	K8s = mockAPI
	defer func() { K8s = originalK8s }()

	expectedArgs := []string{"--kubeconfig=/path/to/config", "get", "se", "-n", "test-namespace"}
	mockAPI.On("ExecuteKubectlCommand", expectedArgs).Return("", assert.AnError)

	wrapper := K8sWrapper{}
	serviceEntries, err := wrapper.GetServiceEntries("/path/to/config", "test-namespace")

	assert.Error(t, err)
	assert.Empty(t, serviceEntries)
	mockAPI.AssertExpectations(t)
}

func TestGetServiceEntries_EmptyOutput(t *testing.T) {
	// Create mock
	mockAPI := &MockK8sAPI{}
	originalK8s := K8s
	K8s = mockAPI
	defer func() { K8s = originalK8s }()

	// Mock kubectl output with only header
	kubectlOutput := `NAME                    HOSTS                   LOCATION   RESOLUTION   AGE`

	expectedArgs := []string{"--kubeconfig=/path/to/config", "get", "se", "-n", "test-namespace"}
	mockAPI.On("ExecuteKubectlCommand", expectedArgs).Return(kubectlOutput, nil)

	wrapper := K8sWrapper{}
	serviceEntries, err := wrapper.GetServiceEntries("/path/to/config", "test-namespace")

	assert.NoError(t, err)
	assert.Empty(t, serviceEntries)
	mockAPI.AssertExpectations(t)
}

// Test error scenarios for other functions
func TestPatchResource_Error(t *testing.T) {
	// Create mock
	mockAPI := &MockK8sAPI{}
	originalK8s := K8s
	K8s = mockAPI
	defer func() { K8s = originalK8s }()

	expectedArgs := []string{"patch", "--kubeconfig=/path/to/config", "-n", "test-namespace", "deployment", "test-deployment", "-p", `{"spec":{"replicas":3}}`, "--type", "json"}
	mockAPI.On("ExecuteKubectlCommand", expectedArgs).Return("", assert.AnError)

	wrapper := K8sWrapper{}
	output, err := wrapper.PatchResource("/path/to/config", "test-namespace", "deployment", "test-deployment", `{"spec":{"replicas":3}}`)

	assert.Error(t, err)
	assert.Equal(t, "", output)
	mockAPI.AssertExpectations(t)
}

func TestScaleDeployment_Error(t *testing.T) {
	// Create mock
	mockAPI := &MockK8sAPI{}
	originalK8s := K8s
	K8s = mockAPI
	defer func() { K8s = originalK8s }()

	expectedArgs := []string{"scale", "--kubeconfig=/path/to/config", "-n", "test-namespace", "deployment/test-deployment", "--replicas=5"}
	mockAPI.On("ExecuteKubectlCommand", expectedArgs).Return("", assert.AnError)

	wrapper := K8sWrapper{}
	output, err := wrapper.ScaleDeployment("/path/to/config", "test-namespace", "test-deployment", 5)

	assert.Error(t, err)
	assert.Equal(t, "", output)
	mockAPI.AssertExpectations(t)
}

func TestRestartContainers_Success(t *testing.T) {
	// Create mock
	mockAPI := &MockK8sAPI{}
	originalK8s := K8s
	K8s = mockAPI
	defer func() { K8s = originalK8s }()

	// Mock running pods
	pods := []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod-1",
				Namespace: "test-namespace",
			},
		},
	}

	// Setup expectations
	mockAPI.On("GetRunningPods", mock.Anything, "test-namespace", "test-app").Return(pods, nil)
	mockAPI.On("RestartContainer", "/path/to/config", "test-namespace", "test-pod-1", "test-container").Return("container restarted", nil)

	wrapper := K8sWrapper{}
	output, err := wrapper.RestartContainers(nil, "/path/to/config", "test-namespace", "test-app", "test-container")

	assert.NoError(t, err)
	assert.Equal(t, "container restarted", output)
	mockAPI.AssertExpectations(t)
}

func TestRestartContainers_GetPodsError(t *testing.T) {
	// Create mock
	mockAPI := &MockK8sAPI{}
	originalK8s := K8s
	K8s = mockAPI
	defer func() { K8s = originalK8s }()

	// Setup expectations for error
	mockAPI.On("GetRunningPods", mock.Anything, "test-namespace", "test-app").Return([]corev1.Pod{}, assert.AnError)

	wrapper := K8sWrapper{}
	output, err := wrapper.RestartContainers(nil, "/path/to/config", "test-namespace", "test-app", "test-container")

	assert.Error(t, err)
	assert.Equal(t, "", output)
	mockAPI.AssertExpectations(t)
}

func TestRestartContainers_NoPods(t *testing.T) {
	// Create mock
	mockAPI := &MockK8sAPI{}
	originalK8s := K8s
	K8s = mockAPI
	defer func() { K8s = originalK8s }()

	// Setup expectations for no pods
	mockAPI.On("GetRunningPods", mock.Anything, "test-namespace", "test-app").Return([]corev1.Pod{}, nil)

	wrapper := K8sWrapper{}
	output, err := wrapper.RestartContainers(nil, "/path/to/config", "test-namespace", "test-app", "test-container")

	assert.NoError(t, err)
	assert.Equal(t, "", output)
	mockAPI.AssertExpectations(t)
}
