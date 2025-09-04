package k8s

import (
	"testing"

	argov1alpha1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	argofake "github.com/argoproj/argo-rollouts/pkg/client/clientset/versioned/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRolloutRestartArgoRollouts(t *testing.T) {
	// Create mock
	mockAPI := &MockK8sAPI{}
	originalK8s := K8s
	K8s = mockAPI
	defer func() { K8s = originalK8s }()

	// Setup expectations
	expectedArgs := []string{"argo", "rollouts", "--kubeconfig=/path/to/config", "-n", "test-namespace", "restart", "test-rollout"}
	mockAPI.On("ExecuteKubectlCommand", expectedArgs).Return("rollout.argoproj.io/test-rollout restarted", nil)

	wrapper := K8sWrapper{}
	output, err := wrapper.RolloutRestartArgoRollouts("/path/to/config", "test-namespace", "test-rollout")

	assert.NoError(t, err)
	assert.Equal(t, "rollout.argoproj.io/test-rollout restarted", output)
	mockAPI.AssertExpectations(t)
}

func TestRolloutRestartArgoRollouts_Error(t *testing.T) {
	// Create mock
	mockAPI := &MockK8sAPI{}
	originalK8s := K8s
	K8s = mockAPI
	defer func() { K8s = originalK8s }()

	// Setup expectations
	expectedArgs := []string{"argo", "rollouts", "--kubeconfig=/path/to/config", "-n", "test-namespace", "restart", "test-rollout"}
	mockAPI.On("ExecuteKubectlCommand", expectedArgs).Return("", assert.AnError)

	wrapper := K8sWrapper{}
	output, err := wrapper.RolloutRestartArgoRollouts("/path/to/config", "test-namespace", "test-rollout")

	assert.Error(t, err)
	assert.Equal(t, "", output)
	mockAPI.AssertExpectations(t)
}

func TestRolloutArogRolloutStatus(t *testing.T) {
	// Create mock
	mockAPI := &MockK8sAPI{}
	originalK8s := K8s
	K8s = mockAPI
	defer func() { K8s = originalK8s }()

	// Setup expectations
	expectedArgs := []string{"argo", "rollouts", "--kubeconfig=/path/to/config", "status", "-n", "test-namespace", "test-rollout", "--watch=false"}
	mockAPI.On("ExecuteKubectlCommand", expectedArgs).Return("Healthy", nil)

	wrapper := K8sWrapper{}
	output, err := wrapper.RolloutArogRolloutStatus("/path/to/config", "test-namespace", "test-rollout")

	assert.NoError(t, err)
	assert.Equal(t, "Healthy", output)
	mockAPI.AssertExpectations(t)
}

func TestRolloutArogRolloutStatus_Error(t *testing.T) {
	// Create mock
	mockAPI := &MockK8sAPI{}
	originalK8s := K8s
	K8s = mockAPI
	defer func() { K8s = originalK8s }()

	// Setup expectations
	expectedArgs := []string{"argo", "rollouts", "--kubeconfig=/path/to/config", "status", "-n", "test-namespace", "test-rollout", "--watch=false"}
	mockAPI.On("ExecuteKubectlCommand", expectedArgs).Return("", assert.AnError)

	wrapper := K8sWrapper{}
	output, err := wrapper.RolloutArogRolloutStatus("/path/to/config", "test-namespace", "test-rollout")

	assert.Error(t, err)
	assert.Equal(t, "", output)
	mockAPI.AssertExpectations(t)
}

func TestGetArgoRollouts_Success(t *testing.T) {
	// Create fake argo clientset
	rollout1 := &argov1alpha1.Rollout{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-rollout-1",
			Namespace: "test-namespace",
		},
	}
	rollout2 := &argov1alpha1.Rollout{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-rollout-2",
			Namespace: "test-namespace",
		},
	}

	fakeArgoClientSet := argofake.NewSimpleClientset(rollout1, rollout2)

	wrapper := K8sWrapper{}
	rolloutList, err := wrapper.GetArgoRollouts(fakeArgoClientSet, "test-namespace")

	assert.NoError(t, err)
	assert.NotNil(t, rolloutList)
	assert.Len(t, rolloutList.Items, 2)
	assert.Equal(t, "test-rollout-1", rolloutList.Items[0].Name)
	assert.Equal(t, "test-rollout-2", rolloutList.Items[1].Name)
}

func TestGetArgoRollouts_EmptyNamespace(t *testing.T) {
	// Create fake argo clientset with no rollouts
	fakeArgoClientSet := argofake.NewSimpleClientset()

	wrapper := K8sWrapper{}
	rolloutList, err := wrapper.GetArgoRollouts(fakeArgoClientSet, "empty-namespace")

	assert.NoError(t, err)
	assert.NotNil(t, rolloutList)
	assert.Len(t, rolloutList.Items, 0)
}

func TestGetArgoRolloutsWithSpecAnnotationFilter_Success(t *testing.T) {
	// Create mock
	mockAPI := &MockK8sAPI{}
	originalK8s := K8s
	K8s = mockAPI
	defer func() { K8s = originalK8s }()

	// Create rollouts with different annotations
	rollout1 := argov1alpha1.Rollout{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-rollout-1",
			Namespace: "test-namespace",
		},
		Spec: argov1alpha1.RolloutSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"app":     "test-app",
						"version": "v1.0",
					},
				},
			},
		},
	}

	rollout2 := argov1alpha1.Rollout{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-rollout-2",
			Namespace: "test-namespace",
		},
		Spec: argov1alpha1.RolloutSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"app":     "other-app",
						"version": "v2.0",
					},
				},
			},
		},
	}

	rolloutList := &argov1alpha1.RolloutList{
		Items: []argov1alpha1.Rollout{rollout1, rollout2},
	}

	// Setup mock expectations
	mockAPI.On("GetArgoRollouts", mock.Anything, "test-namespace").Return(rolloutList, nil)

	wrapper := K8sWrapper{}
	filterAnnotations := map[string]string{"app": "test-app"}
	filteredRollouts, err := wrapper.GetArgoRolloutsWithSpecAnnotationFilter(nil, "test-namespace", filterAnnotations)

	assert.NoError(t, err)
	assert.Len(t, filteredRollouts, 1)
	assert.Equal(t, "test-rollout-1", filteredRollouts[0])
	mockAPI.AssertExpectations(t)
}

func TestGetArgoRolloutsWithSpecAnnotationFilter_Error(t *testing.T) {
	// Create mock
	mockAPI := &MockK8sAPI{}
	originalK8s := K8s
	K8s = mockAPI
	defer func() { K8s = originalK8s }()

	// Setup mock expectations for error
	mockAPI.On("GetArgoRollouts", mock.Anything, "test-namespace").Return(&argov1alpha1.RolloutList{}, assert.AnError)

	wrapper := K8sWrapper{}
	filterAnnotations := map[string]string{"app": "test-app"}
	filteredRollouts, err := wrapper.GetArgoRolloutsWithSpecAnnotationFilter(nil, "test-namespace", filterAnnotations)

	assert.Error(t, err)
	assert.Empty(t, filteredRollouts)
	mockAPI.AssertExpectations(t)
}

func TestGetArgoRolloutsWithSpecAnnotationFilter_NoMatches(t *testing.T) {
	// Create mock
	mockAPI := &MockK8sAPI{}
	originalK8s := K8s
	K8s = mockAPI
	defer func() { K8s = originalK8s }()

	// Create rollout without matching annotations
	rollout1 := argov1alpha1.Rollout{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-rollout-1",
			Namespace: "test-namespace",
		},
		Spec: argov1alpha1.RolloutSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"app":     "other-app",
						"version": "v1.0",
					},
				},
			},
		},
	}

	rolloutList := &argov1alpha1.RolloutList{
		Items: []argov1alpha1.Rollout{rollout1},
	}

	// Setup mock expectations
	mockAPI.On("GetArgoRollouts", mock.Anything, "test-namespace").Return(rolloutList, nil)

	wrapper := K8sWrapper{}
	filterAnnotations := map[string]string{"app": "test-app"}
	filteredRollouts, err := wrapper.GetArgoRolloutsWithSpecAnnotationFilter(nil, "test-namespace", filterAnnotations)

	assert.NoError(t, err)
	assert.Empty(t, filteredRollouts)
	mockAPI.AssertExpectations(t)
}
