package k8s

import (
	"context"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"keikoproj.intuit.com/Flippy/pkg/k8s-utils/utils"
)

func (K8sWrapper) GetNamespaces(clientset kubernetes.Interface) (*v1.NamespaceList, error) {
	namespaceClient := clientset.CoreV1().Namespaces()
	namespaces, err := namespaceClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return namespaces, nil
}

func (K8sWrapper) GetNamespaceWithLabelFilter(clientset kubernetes.Interface, labels map[string]string) ([]string, error) {

	labelSelectorFilter := ""

	for labelName, labelValue := range labels {
		if len(labelSelectorFilter) > 0 {
			labelSelectorFilter += ", "
		}
		labelSelectorFilter = labelName + "=" + labelValue
	}

	filteredNamespaces := make([]string, 0)

	namespaceClient := clientset.CoreV1().Namespaces()
	namespaces, err := namespaceClient.List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelectorFilter,
	})

	if err != nil {
		return filteredNamespaces, nil
	}

	for _, namespace := range namespaces.Items {
		if utils.IsStringMapSubset(namespace.ObjectMeta.Labels, labels) {
			filteredNamespaces = append(filteredNamespaces, namespace.Name)
		}
	}
	return filteredNamespaces, nil
}
