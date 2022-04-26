package k8s

import (
	"context"
	"k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func (K8sWrapper) CreateNetworkPolicy(clientset kubernetes.Interface, networkpolicy *v1.NetworkPolicy, namespace string) error {
	networkpolicyClient := clientset.NetworkingV1().NetworkPolicies(namespace)
	_, err := networkpolicyClient.Create(context.TODO(), networkpolicy, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

//func (K8sWrapper) DeleteNetworkPolicy(clientset kubernetes.Interface, namespace string, networkpolicyname string) error {
//	networkpolicyClient := clientset.NetworkingV1().NetworkPolicies(namespace)
//	err := networkpolicyClient.Delete(context.TODO(), networkpolicyname, metav1.DeleteOptions{})
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func (K8sWrapper) CreateNetworkPolicyWithCustomization(clientset kubernetes.Interface, networkpolicy *v1.NetworkPolicy, namespace string, podSelector map[string]string) error {
//
//	if podSelector != nil {
//		networkpolicy.Spec.PodSelector.MatchLabels = podSelector
//	} else {
//		networkpolicy.Spec.PodSelector = metav1.LabelSelector{}
//	}
//	return K8s.CreateNetworkPolicy(clientset, networkpolicy, namespace)
//}
