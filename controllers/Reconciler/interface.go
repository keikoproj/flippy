package Reconciler

import (
	crdv1 "github.com/keikoproj/flippy/api/v1"
	"github.com/keikoproj/flippy/pkg/common"
	"github.com/keikoproj/flippy/pkg/k8s-utils/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type ReconcilerInterface interface {
	Handle(crdv1.FlippyConfig, k8s.ClientSet, k8s.K8sAPI) error
	IsProxyVersionChange([]string) bool
	IsPreConditionSatisfied(k8s.K8sAPI, kubernetes.Interface, crdv1.FlippyCondition, int) bool

	ProcessRestart(k8s k8s.K8sAPI, clientset k8s.ClientSet, config crdv1.FlippyConfig) error
	FilterNameSpaceNeedAttention(clientset k8s.ClientSet, namespaces []string, config crdv1.FlippyConfig, k8sapi k8s.K8sAPI) []common.RestartObjects

	IsAnyPodContainsContainers(podList []corev1.Pod, containers []string) bool
	IsPodContainContainers(pod corev1.Pod, containers []string) []corev1.ContainerStatus
	IsAnyPodRunningWithProvidedDockerImage(podList []corev1.Pod, dockerImages []string, containers []string) bool

	ProcessNamespaceRestarts(k8s k8s.K8sAPI, restartObjects []common.RestartObjects)
}

type ReconcilerWrapper struct{}

var Process ReconcilerInterface = ReconcilerWrapper{}
