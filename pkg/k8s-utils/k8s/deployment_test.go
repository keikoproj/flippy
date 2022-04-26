package k8s

import (
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
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
