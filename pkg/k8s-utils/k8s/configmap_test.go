package k8s

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"reflect"
	"testing"
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
