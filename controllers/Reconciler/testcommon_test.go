package Reconciler

import (
	"errors"
	argov1alpha1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	argo "github.com/argoproj/argo-rollouts/pkg/client/clientset/versioned"
	flippyv1 "github.com/keikoproj/flippy/api/v1"
	"github.com/keikoproj/flippy/pkg/common"
	"github.com/keikoproj/flippy/pkg/k8s-utils/k8s"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metaapiv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"time"
)

type K8sWrapperFakeSuccess struct{}

var fakeK8sApi k8s.K8sAPI = K8sWrapperFakeSuccess{}

var HappyPreCondition = "HappyPreCondition"
var NotHappyPreCondition = "NotHappyPreCondition"

var HappyImageList = []string{"docker.com/intuit/HappyImage:latest", "docker.com/intuit/HappyImage:dev"}

func (k K8sWrapperFakeSuccess) GetDeployment(clientset kubernetes.Interface, namespace string, deploymentName string) (appsv1.Deployment, error) {
	mockDeployment := appsv1.Deployment{

		TypeMeta: metaapiv1.TypeMeta{
			Kind: "Deployment",
		},
		ObjectMeta: metaapiv1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespace,
		},
		//Spec:       ,
		Status: appsv1.DeploymentStatus{
			ObservedGeneration:  0,
			Replicas:            1,
			UpdatedReplicas:     1,
			ReadyReplicas:       1,
			AvailableReplicas:   1,
			UnavailableReplicas: 0,
			Conditions:          nil,
			CollisionCount:      nil,
		},
	}

	if namespace == NotHappyPreCondition && deploymentName == NotHappyPreCondition {
		mockDeployment.Namespace = NotHappyPreCondition
		mockDeployment.Name = NotHappyPreCondition
		mockDeployment.Status.AvailableReplicas = 0
		mockDeployment.Status.UnavailableReplicas = 2
		mockDeployment.Status.ReadyReplicas = 0
	}

	return mockDeployment, nil
}

func (k K8sWrapperFakeSuccess) RolloutDeploymentStatus(kubeconfigpath string, namespace string, deploymentName string) (string, error) {

	if namespace == HappyPreCondition && deploymentName == HappyPreCondition {
		return HappyPreCondition + " successfully rolled out", nil
	} else if namespace == NotHappyPreCondition && deploymentName == NotHappyPreCondition {
		return NotHappyPreCondition, nil
	} else {
		return NotHappyPreCondition, errors.New(NotHappyPreCondition)
	}
}

func (k K8sWrapperFakeSuccess) GetNamespaces(clientset kubernetes.Interface) (*v1.NamespaceList, error) {
	panic("implement me")
}

func (k K8sWrapperFakeSuccess) GetNamespaceWithLabelFilter(clientset kubernetes.Interface, labels map[string]string) ([]string, error) {
	if labels[NotHappyPreCondition] == NotHappyPreCondition {
		return []string{}, errors.New(NotHappyPreCondition)
	}
	return []string{HappyPreCondition}, nil
}

func (k K8sWrapperFakeSuccess) GetDeploymentWithSpecAnnotationFilter(clientset kubernetes.Interface, namespace string, annotationName map[string]string) ([]string, error) {
	if namespace == NotHappyPreCondition {
		return []string{NotHappyPreCondition}, nil
	}
	return []string{HappyPreCondition}, nil
}

func (k K8sWrapperFakeSuccess) ReadConfigMap(clientset kubernetes.Interface, namespace string, configMapName string) (*v1.ConfigMap, error) {
	panic("implement me")
}

func (k K8sWrapperFakeSuccess) ReadConfigMapData(clientset kubernetes.Interface, namespace string, configMapName string) (map[string]string, error) {
	panic("implement me")
}

func (k K8sWrapperFakeSuccess) GetAllPodsInNamespace(clientset kubernetes.Interface, namespace string) ([]corev1.Pod, error) {
	var fakeContainerStatus []corev1.ContainerStatus
	var fakePods []corev1.Pod

	fakeContainerStatus = append(fakeContainerStatus, corev1.ContainerStatus{
		Name:        "istio-proxy",
		Image:       HappyImageList[0],
		ContainerID: "SampleContainer",
	})

	fakePods = append(fakePods, corev1.Pod{
		Status: corev1.PodStatus{
			ContainerStatuses: fakeContainerStatus,
		},
	})
	return fakePods, nil
}

func (k K8sWrapperFakeSuccess) GetArgoRollouts(clientset argo.Interface, namespace string) (*argov1alpha1.RolloutList, error) {
	panic("implement me")
}

func (k K8sWrapperFakeSuccess) GetArgoRolloutsWithSpecAnnotationFilter(clientset argo.Interface, namespace string, mapObject map[string]string) ([]string, error) {
	return []string{HappyPreCondition}, nil
}

func (k K8sWrapperFakeSuccess) RolloutRestartArgoRollouts(kubeconfigpath string, namespace string, deploymentName string) (string, error) {
	return "Success", nil
}

func (k K8sWrapperFakeSuccess) RolloutArogRolloutStatus(kubeconfigpath string, namespace string, deploymentName string) (string, error) {
	if namespace == HappyPreCondition && deploymentName == HappyPreCondition {
		return HappyPreCondition + " Healthy", nil
	} else if namespace == NotHappyPreCondition && deploymentName == NotHappyPreCondition {
		return NotHappyPreCondition, nil
	} else {
		return NotHappyPreCondition, errors.New(NotHappyPreCondition)
	}
}

func (k K8sWrapperFakeSuccess) PatchResource(kubeconfigpath string, namespace string, resource string, resourceName string, patchJson string) error {
	panic("implement me")
}

func (k K8sWrapperFakeSuccess) ScaleDeployment(kubeconfigpath string, namespace string, deploymentname string, scale int) error {
	panic("implement me")
}

func (k K8sWrapperFakeSuccess) ApplyYaml(kubeconfigpath string, yamlFilePath string) error {
	panic("implement me")
}

func (k K8sWrapperFakeSuccess) DeleteYaml(kubeconfigpath string, yamlFilePath string) error {
	panic("implement me")
}

func (k K8sWrapperFakeSuccess) GetServiceEntries(kubeconfigpath string, namespace string) (map[string]string, error) {
	panic("implement me")
}

func (k K8sWrapperFakeSuccess) RestartContainer(kubeconfigpath string, namespace string, podname string, containerName string) (string, error) {
	panic("implement me")
}

func (k K8sWrapperFakeSuccess) RolloutRestartDeployment(kubeconfigpath string, namespace string, deploymentName string) (string, error) {
	return "Success", nil
}

func (k K8sWrapperFakeSuccess) GetRunningPods(clientset kubernetes.Interface, namespace string, podNameContains string) ([]corev1.Pod, error) {
	panic("implement me")
}

func (k K8sWrapperFakeSuccess) DeletePod(clientset kubernetes.Interface, namespace string, podName string) error {
	panic("implement me")
}

func (k K8sWrapperFakeSuccess) DeletePodsWithRetry(clientset kubernetes.Interface, namespace string, podNameContains string) error {
	panic("implement me")
}

func (k K8sWrapperFakeSuccess) DeletePods(clientset kubernetes.Interface, namespace string, podNameContains string) error {
	panic("implement me")
}

func (k K8sWrapperFakeSuccess) GetPodLogs(clientset kubernetes.Interface, podName string, namespace string, container string, fromTime time.Time) (string, error) {
	panic("implement me")
}

func (k K8sWrapperFakeSuccess) RestartContainers(clientset kubernetes.Interface, kubeconfigpath string, namespace string, podnamecontains string, containerName string) (string, error) {
	panic("implement me")
}

func (k K8sWrapperFakeSuccess) GetLogFromFirstPod(clientset kubernetes.Interface, namespace string, podNameContains string, containerName string, logsFromTime time.Time) (string, error) {
	panic("implement me")
}

func (k K8sWrapperFakeSuccess) CreateNetworkPolicy(clientset kubernetes.Interface, networkpolicy *networkingv1.NetworkPolicy, namespace string) error {
	panic("implement me")
}

func (k K8sWrapperFakeSuccess) DeleteNetworkPolicy(clientset kubernetes.Interface, namespace string, networkpolicyname string) error {
	panic("implement me")
}

func (k K8sWrapperFakeSuccess) CreateNetworkPolicyWithCustomization(clientset kubernetes.Interface, networkpolicy *networkingv1.NetworkPolicy, namespace string, podSelector map[string]string) error {
	panic("implement me")
}

func (k K8sWrapperFakeSuccess) GetNameSpaces(clientset kubernetes.Interface) (*corev1.NamespaceList, error) {
	panic("implement me")
}

func (k K8sWrapperFakeSuccess) GetNameSpaceWithLabelFilter(clientset kubernetes.Interface, labelName string, labelValue string) ([]string, error) {
	return []string{HappyPreCondition}, nil
}

func (k K8sWrapperFakeSuccess) CreateJob(clientset kubernetes.Interface, job *batchv1.Job, namespace string) error {
	panic("implement me")
}

func (k K8sWrapperFakeSuccess) DeleteJob(clientset kubernetes.Interface, namespace string, jobName string) error {
	panic("implement me")
}

func (k K8sWrapperFakeSuccess) DeleteJobWithPods(clientset kubernetes.Interface, namespace string, jobName string) error {
	panic("implement me")
}

func (k K8sWrapperFakeSuccess) DeleteDeployment(clientset kubernetes.Interface, namespace string, deploymentName string) error {
	panic("implement me")
}

func (k K8sWrapperFakeSuccess) GetDeployments(clientset kubernetes.Interface, namespace string) (*appsv1.DeploymentList, error) {
	panic("implement me")
}

func BuildTestK8SClientSet() k8s.ClientSet {
	testClientSet := k8s.ClientSet{
		K8sClientSet:         fake.NewSimpleClientset(),
		ArgoRolloutClientSet: nil,
	}

	return testClientSet
}

func BuildTestFlippyStatusCheckConfig() flippyv1.StatusCheckConfig {
	testStatusConfig := flippyv1.StatusCheckConfig{
		CheckStatus:   false,
		MaxRetry:      0,
		RetryDuration: 0,
	}

	return testStatusConfig
}

func BuildTestFlippyConfig() flippyv1.FlippyConfig {

	NamespaceLabel := make(map[string]string)
	NamespaceLabel["sidecar.istio.io/inject"] = "true"

	testStatusConfig := BuildTestFlippyStatusCheckConfig()

	var happyPostFilterRestartsTests []flippyv1.FlippyCondition
	happyPostFilterRestartsTests = append(happyPostFilterRestartsTests, flippyv1.FlippyCondition{
		K8S: flippyv1.K8S{
			Type:      common.DEPLOYMENT,
			Name:      HappyPreCondition,
			Namespace: HappyPreCondition,
		},
		Status:            "",
		StatusCheckConfig: testStatusConfig,
	})

	happyPostFilterRestartsTests = append(happyPostFilterRestartsTests, flippyv1.FlippyCondition{
		K8S: flippyv1.K8S{
			Type:      common.ARGO_ROLLOUT,
			Name:      HappyPreCondition,
			Namespace: HappyPreCondition,
		},
		Status:            "",
		StatusCheckConfig: testStatusConfig,
	})

	var restartObjectTests []flippyv1.RestartObject
	restartObjectTests = append(restartObjectTests, flippyv1.RestartObject{
		Type:              common.DEPLOYMENT,
		StatusCheckConfig: testStatusConfig,
	})
	restartObjectTests = append(restartObjectTests, flippyv1.RestartObject{
		Type:              common.ARGO_ROLLOUT,
		StatusCheckConfig: testStatusConfig,
	})

	flippyConfigTest := flippyv1.FlippyConfig{
		TypeMeta:   metaapiv1.TypeMeta{},
		ObjectMeta: metaapiv1.ObjectMeta{},
		Spec: flippyv1.FlippyConfigSpec{
			ImageList:     HappyImageList,
			Preconditions: happyPostFilterRestartsTests,
			ProcessFilter: flippyv1.ProcessFilter{
				PodLabels:         nil,
				NamespaceLabels:   NamespaceLabel,
				Annotations:       nil,
				Containers:        []string{"istio-proxy"},
				PreProcessRestart: flippyv1.K8S{},
			},
			RestartObjects:     restartObjectTests,
			PostFilterRestarts: happyPostFilterRestartsTests,
		},
		Status: flippyv1.FlippyConfigStatus{},
	}

	return flippyConfigTest
}

func IsRestartObjectSame(objs1 []common.RestartObjects, objs2 []common.RestartObjects) bool {
	if len(objs1) == len(objs2) {
		match := 0
		for i := 0; i < len(objs1); i++ {
			for j := 0; j < len(objs2); j++ {
				if objs1[i].Type == objs2[j].Type &&
					objs1[i].RestartConfig == objs2[j].RestartConfig &&
					len(objs1[i].NamespaceObjects) == len(objs2[j].NamespaceObjects) {
					match++
					break
				}
			}
		}
		if match == len(objs1) {
			return true
		}
	} else {
		return false
	}

	return false
}
