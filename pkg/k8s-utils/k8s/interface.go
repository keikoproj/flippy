package k8s

import (
	argov1alpha1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	argo "github.com/argoproj/argo-rollouts/pkg/client/clientset/versioned"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

type K8sAPI interface {
	PatchResource(kubeconfigpath string, namespace string, resource string, resourceName string, patchJson string) error
	ScaleDeployment(kubeconfigpath string, namespace string, deploymentname string, scale int) error
	ApplyYaml(kubeconfigpath string, yamlFilePath string) error
	DeleteYaml(kubeconfigpath string, yamlFilePath string) error
	GetServiceEntries(kubeconfigpath string, namespace string) (map[string]string, error)
	RestartContainer(kubeconfigpath string, namespace string, podname string, containerName string) (string, error)
	RolloutRestartDeployment(kubeconfigpath string, namespace string, deploymentName string) (string, error)
	RolloutDeploymentStatus(kubeconfigpath string, namespace string, deploymentName string) (string, error)

	GetRunningPods(clientset kubernetes.Interface, namespace string, podNameContains string) ([]corev1.Pod, error)
	GetAllPodsInNamespace(clientset kubernetes.Interface, namespace string) ([]corev1.Pod, error)
	DeletePod(clientset kubernetes.Interface, namespace string, podName string) error
	DeletePodsWithoutRetry(clientset kubernetes.Interface, namespace string, podNameContains string) error
	DeletePods(clientset kubernetes.Interface, namespace string, podNameContains string) error
	GetPodLogs(clientset kubernetes.Interface, podName string, namespace string, container string, fromTime time.Time) (string, error)
	RestartContainers(clientset kubernetes.Interface, kubeconfigpath string, namespace string, podnamecontains string, containerName string) (string, error)
	GetLogFromFirstPod(clientset kubernetes.Interface, namespace string, podNameContains string, containerName string, logsFromTime time.Time) (string, error)

	CreateNetworkPolicy(clientset kubernetes.Interface, networkpolicy *networkingv1.NetworkPolicy, namespace string) error
	DeleteNetworkPolicy(clientset kubernetes.Interface, namespace string, networkpolicyname string) error
	CreateNetworkPolicyWithCustomization(clientset kubernetes.Interface, networkpolicy *networkingv1.NetworkPolicy, namespace string, podSelector map[string]string) error

	GetNamespaces(clientset kubernetes.Interface) (*corev1.NamespaceList, error)
	GetNamespaceWithLabelFilter(clientset kubernetes.Interface, labels map[string]string) ([]string, error)

	CreateJob(clientset kubernetes.Interface, job *batchv1.Job, namespace string) error
	DeleteJob(clientset kubernetes.Interface, namespace string, jobName string) error
	DeleteJobWithPods(clientset kubernetes.Interface, namespace string, jobName string) error

	DeleteDeployment(clientset kubernetes.Interface, namespace string, deploymentName string) error
	GetDeployments(clientset kubernetes.Interface, namespace string) (*appsv1.DeploymentList, error)
	GetDeployment(clientset kubernetes.Interface, namespace string, deploymentName string) (appsv1.Deployment, error)
	GetDeploymentWithSpecAnnotationFilter(clientset kubernetes.Interface, namespace string, annotationName map[string]string) ([]string, error)

	//GetArgoClient()
	GetArgoRollouts(clientset argo.Interface, namespace string) (*argov1alpha1.RolloutList, error)
	GetArgoRolloutsWithSpecAnnotationFilter(clientset argo.Interface, namespace string, annotations map[string]string) ([]string, error)
	RolloutRestartArgoRollouts(kubeconfigpath string, namespace string, argoRolloutName string) (string, error)
	RolloutArogRolloutStatus(kubeconfigpath string, namespace string, argoRolloutName string) (string, error)

	ReadConfigMap(clientset kubernetes.Interface, namespace string, configMapName string) (*corev1.ConfigMap, error)
	ReadConfigMapData(clientset kubernetes.Interface, namespace string, configMapName string) (map[string]string, error)
}

type K8sWrapper struct{}

var K8s K8sAPI = K8sWrapper{}

type ClientSet struct {
	K8sClientSet         kubernetes.Interface
	ArgoRolloutClientSet argo.Interface
}
