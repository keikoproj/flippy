package k8s

import (
	"testing"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetDeployment(t *testing.T) {

	fakeClientSet := fake.NewSimpleClientset(&v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "TestDeployment",
			Namespace:   "TestNamespace",
			Annotations: map[string]string{},
		},
	})

	k8 := K8sWrapper{}

	deployment, err := k8.GetDeployment(fakeClientSet, "TestNamespace", "TestDeployment")
	if err != nil {
		t.Errorf("GetDeployment test failed, expected no error. Actual error = %v", err)
		return
	}

	if deployment.Name != "TestDeployment" {
		t.Errorf("GetDeployment test failed, expected deployment name %v, but got %v", "TestDeployment", deployment.Name)
		return
	}
}

func TestDeleteDeployment(t *testing.T) {
	fakeClientSet := fake.NewSimpleClientset(&v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "TestDeployment",
			Namespace:   "TestNamespace",
			Annotations: map[string]string{},
		},
	})

	k8 := K8sWrapper{}

	err := k8.DeleteDeployment(fakeClientSet, "TestNamespace", "TestDeployment")

	if err != nil {
		t.Errorf("DeleteDeployment test failed, expected no error. Actual error = %v", err)
		return
	}
}

func TestGetDeploymentWithSpecAnnotationFilter(t *testing.T) {
	annotations := map[string]string{"istio-injection": "true", "identifier": "test"}

	fakeClientSet := fake.NewSimpleClientset(
		&v1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "TestDeploymentWithAnnotations",
				Namespace: "TestNamespace",
			},
			Spec: v1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: annotations,
					},
				},
			},
		},
		&v1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "TestDeploymentWithAnnotation1",
				Namespace: "TestNamespace",
			},
		})

	k8 := K8sWrapper{}

	deployments, err := k8.GetDeploymentWithSpecAnnotationFilter(fakeClientSet, "TestNamespace", map[string]string{"istio-injection": "true", "identifier": "test"})

	if err != nil {
		t.Errorf("GetDeploymentWithSpecAnnotationFilter test failed, expected no error. Actual error = %v", err)
		return
	}

	if deployments[0] != "TestDeploymentWithAnnotations" {
		t.Errorf("GetDeploymentWithSpecAnnotationFilter test failed, expected deployment name %v, but got %v", "TestDeploymentWithAnnotations", deployments[0])
		return
	}
}

// TestGetDeployment_ErrorPaths tests error handling in GetDeployment
func TestGetDeployment_ErrorPaths(t *testing.T) {
	tests := []struct {
		name           string
		setupClientset func() *fake.Clientset
		namespace      string
		deploymentName string
		expectError    bool
		expectEmpty    bool
		description    string
	}{
		{
			name: "NonExistentDeployment",
			setupClientset: func() *fake.Clientset {
				// Create clientset with different deployment
				return fake.NewSimpleClientset(&v1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "OtherDeployment",
						Namespace: "TestNamespace",
					},
				})
			},
			namespace:      "TestNamespace",
			deploymentName: "NonExistentDeployment",
			expectError:    false,
			expectEmpty:    true,
			description:    "Should return empty deployment when deployment not found",
		},
		{
			name: "EmptyNamespace",
			setupClientset: func() *fake.Clientset {
				// Create empty clientset
				return fake.NewSimpleClientset()
			},
			namespace:      "TestNamespace",
			deploymentName: "AnyDeployment",
			expectError:    false,
			expectEmpty:    true,
			description:    "Should return empty deployment when namespace has no deployments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientset := tt.setupClientset()
			k8 := K8sWrapper{}

			deployment, err := k8.GetDeployment(clientset, tt.namespace, tt.deploymentName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none: %s", tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got %v: %s", err, tt.description)
				}
				if tt.expectEmpty {
					if deployment.Name != "" {
						t.Errorf("Expected empty deployment but got %v: %s", deployment.Name, tt.description)
					}
				}
			}
		})
	}
}

// TestDeleteDeployment_ErrorPath tests error handling in DeleteDeployment
func TestDeleteDeployment_ErrorPath(t *testing.T) {
	// Create empty clientset - no deployments
	fakeClientSet := fake.NewSimpleClientset()

	k8 := K8sWrapper{}

	// Try to delete non-existent deployment
	err := k8.DeleteDeployment(fakeClientSet, "TestNamespace", "NonExistentDeployment")

	// Should return error
	if err == nil {
		t.Errorf("Expected error when deleting non-existent deployment, but got none")
		return
	}
}
