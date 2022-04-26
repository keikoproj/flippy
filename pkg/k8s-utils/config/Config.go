package config

import (
	argo "github.com/argoproj/argo-rollouts/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func Setup(kubeConfigPath string) (*rest.Config, *kubernetes.Clientset, error) {

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	return config, clientset, err
}

func SetupInCluster() (*rest.Config, *kubernetes.Clientset, error) {

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	return config, clientset, err
}

func SetupArgoRollout(kubeConfigPath string) (*rest.Config, *argo.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, nil, err
	}
	clientset, err := argo.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	return config, clientset, err
}

func SetupInClusterArgoRollout() (*rest.Config, *argo.Clientset, error) {

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, nil, err
	}
	clientset, err := argo.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	return config, clientset, err
}
