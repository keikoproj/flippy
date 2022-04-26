package k8s

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestGetNamespaces(t *testing.T) {

	k8 := K8sWrapper{}

	fakeClientSet := fake.NewSimpleClientset(
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "TestNamespace",
			},
		})

	namespaces, err := k8.GetNamespaces(fakeClientSet)
	if err != nil {
		t.Errorf("GetNamespaces test failed, expected no error. Actual error = %v", err)
		return
	}

	if len(namespaces.Items) != 1 {
		t.Errorf("Expected one namespace in response list, but got %v", len(namespaces.Items))
		return
	}
}

func TestGetNamespaceWithLabelFilter(t *testing.T) {
	k8 := K8sWrapper{}

	labels := map[string]string{"label1": "key1", "foo": "bar"}
	fakeClientSet := fake.NewSimpleClientset(
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "TestNamespace",
				Labels: labels,
			},
		}, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "TestNamespace1",
				Labels: map[string]string{"label1": "key1"},
			},
		})

	namespaces, err := k8.GetNamespaceWithLabelFilter(fakeClientSet, labels)
	if err != nil {
		t.Errorf("GetNamespaces test failed, expected no error. Actual error = %v", err)
		return
	}

	if len(namespaces) != 1 && namespaces[0] != "TestNamespace" {
		t.Errorf("Expected one namespace with name %v in response, but got %v with name %v", "TestNamespace", len(namespaces), namespaces[0])
		return
	}
}
