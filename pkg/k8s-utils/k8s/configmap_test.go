package k8s

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestReadConfigMapData(t *testing.T) {

	data := map[string]string{"data_key": "This is test data value"}

	fakeClientSet := fake.NewSimpleClientset(&corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "TestConfigMap",
			Namespace: "TestNamespace",
		},
		Data: data,
	})

	k8 := K8sWrapper{}

	confgiMapData, err := k8.ReadConfigMapData(fakeClientSet, "TestNamespace", "TestConfigMap")
	if err != nil {
		t.Errorf("ReadConfigMapData test failed, expected no error. Actual error = %v", err)
		return
	}

	if !reflect.DeepEqual(confgiMapData, data) {
		t.Errorf("ReadConfigMapData test failed, expected configmap data %v, but got %v", data, confgiMapData)
		return
	}
}

// TestReadConfigMapData_ErrorPath tests error handling in ReadConfigMapData
func TestReadConfigMapData_ErrorPath(t *testing.T) {
	// Create empty clientset - no configmaps
	fakeClientSet := fake.NewSimpleClientset()

	k8 := K8sWrapper{}

	// Try to read non-existent configmap
	configMapData, err := k8.ReadConfigMapData(fakeClientSet, "TestNamespace", "NonExistentConfigMap")

	// Should return error
	if err == nil {
		t.Errorf("Expected error when reading non-existent configmap, but got none")
		return
	}

	// Should return nil data
	if configMapData != nil {
		t.Errorf("Expected nil configmap data on error, but got %v", configMapData)
		return
	}
}
