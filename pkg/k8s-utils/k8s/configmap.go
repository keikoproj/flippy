package k8s

import (
	"context"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func (K8sWrapper) ReadConfigMap(clientset kubernetes.Interface, namespace string, configMapName string) (*v1.ConfigMap, error) {

	configMapClient := clientset.CoreV1().ConfigMaps(namespace)

	configMap, err := configMapClient.Get(context.TODO(), configMapName, metav1.GetOptions{})

	return configMap, err
}

func (K8sWrapper) ReadConfigMapData(clientset kubernetes.Interface, namespace string, configMapName string) (map[string]string, error) {
	configMap, err := K8s.ReadConfigMap(clientset, namespace, configMapName)
	if err != nil {
		return nil, err
	}
	return configMap.Data, nil
}
