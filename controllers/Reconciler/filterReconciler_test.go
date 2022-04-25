package Reconciler

import (
	v1 "github.com/keikoproj/flippy/api/v1"
	"github.com/keikoproj/flippy/pkg/common"
	"github.com/keikoproj/flippy/pkg/k8s-utils/k8s"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestReconcilerWrapper_IsAnyPodContainsContainers(t *testing.T) {

	fakePods := []corev1.Pod{
		{
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{Name: "istio-proxy", Image: "foo_image", ContainerID: "foo"},
				},
			},
		},
		{
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{Name: "Awesome", Image: "foo_image", ContainerID: "zoo"},
				},
			},
		},
	}

	type testargs struct {
		podList    []corev1.Pod
		containers []string
	}
	tests := []struct {
		name string
		args testargs
		want bool
	}{
		{"ContainerPresent", testargs{podList: fakePods, containers: []string{"istio-proxy"}}, true},
		{"ContainerNotPresent", testargs{podList: fakePods, containers: []string{"envoy"}}, false},
		{"EmptyContainerList", testargs{podList: fakePods, containers: []string{}}, false},
		{"EmptyContainerAndPod", testargs{podList: []corev1.Pod{}, containers: []string{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := ReconcilerWrapper{}
			if got := re.IsAnyPodContainsContainers(tt.args.podList, tt.args.containers); got != tt.want {
				t.Errorf("IsAnyPodContainsContainers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReconcilerWrapper_IsAnyPodRunningWithProvidedDockerImage(t *testing.T) {

	fakePods := []corev1.Pod{
		{
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{Name: "istio-proxy", Image: "foo_image", ContainerID: "foo"},
				},
			},
		},
		{
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{Name: "Awesome", Image: "zoo_image", ContainerID: "zoo"},
				},
			},
		},
	}

	type testargs struct {
		podList      []corev1.Pod
		dockerImages []string
		containers   []string
	}
	tests := []struct {
		name string
		args testargs
		want bool
	}{
		{"ContainerAndImageFound", testargs{podList: fakePods, dockerImages: []string{"foo_image"}, containers: []string{"istio-proxy"}}, true},
		{"ContainerFoundButNoImageFound", testargs{podList: fakePods, dockerImages: []string{"doo_image"}, containers: []string{"istio-proxy"}}, false},
		{"NoContainerFoundAndNoImageFound", testargs{podList: fakePods, dockerImages: []string{"doo_image"}, containers: []string{"envoy"}}, false},
		{"EmptyContainer", testargs{podList: fakePods, dockerImages: []string{"doo_image"}, containers: []string{}}, false},
		{"EmptyContainerAndImage", testargs{podList: fakePods, dockerImages: []string{}, containers: []string{}}, false},
		{"EmptyContainerAndImageAndPod", testargs{podList: []corev1.Pod{}, dockerImages: []string{}, containers: []string{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := ReconcilerWrapper{}
			if got := re.IsAnyPodRunningWithProvidedDockerImage(tt.args.podList, tt.args.dockerImages, tt.args.containers); got != tt.want {
				t.Errorf("IsAnyPodRunningWithProvidedDockerImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReconcilerWrapper_FilterNameSpaceNeedAttention(t *testing.T) {
	type args struct {
		clientset  k8s.ClientSet
		namespaces []string
		config     v1.FlippyConfig
		k8sapi     k8s.K8sAPI
	}

	testArg1 := args{
		clientset:  BuildTestK8SClientSet(),
		namespaces: []string{HappyPreCondition, NotHappyPreCondition},
		config:     BuildTestFlippyConfig(),
		k8sapi:     fakeK8sApi,
	}

	testArg2 := testArg1

	happyNamespace := make(map[string][]string)
	happyNamespace[HappyPreCondition] = []string{HappyPreCondition}

	notHappyNamespace := make(map[string][]string)
	notHappyNamespace[NotHappyPreCondition] = []string{NotHappyPreCondition}

	expectedRestartObjectsNonMatchingPodImage := []common.RestartObjects{}

	expectedRestartObjectsNonMatchingPodImage = append(expectedRestartObjectsNonMatchingPodImage, common.RestartObjects{
		Type:             common.DEPLOYMENT,
		NamespaceObjects: happyNamespace,
		RestartConfig:    BuildTestFlippyStatusCheckConfig(),
	})

	expectedRestartObjectsNonMatchingPodImage = append(expectedRestartObjectsNonMatchingPodImage, common.RestartObjects{
		Type:             common.ARGO_ROLLOUT,
		NamespaceObjects: happyNamespace,
		RestartConfig:    BuildTestFlippyStatusCheckConfig(),
	})

	expectedRestartObjectsNonMatchingPodImage = append(expectedRestartObjectsNonMatchingPodImage, common.RestartObjects{
		Type:             common.DEPLOYMENT,
		NamespaceObjects: notHappyNamespace,
		RestartConfig:    BuildTestFlippyStatusCheckConfig(),
	})

	expectedRestartObjectsNonMatchingPodImage = append(expectedRestartObjectsNonMatchingPodImage, common.RestartObjects{
		Type:             common.ARGO_ROLLOUT,
		NamespaceObjects: notHappyNamespace,
		RestartConfig:    BuildTestFlippyStatusCheckConfig(),
	})

	expectedRestartObjectsMatchingPodImage := []common.RestartObjects{}

	for _, obj := range expectedRestartObjectsNonMatchingPodImage {
		obj.NamespaceObjects = nil
		expectedRestartObjectsMatchingPodImage = append(expectedRestartObjectsMatchingPodImage, obj)
	}

	testArg2.config.Spec.ImageList = []string{"foo", "boo"}

	tests := []struct {
		name string
		args args
		want []common.RestartObjects
	}{
		{"NonMatchingPodImages", testArg2, expectedRestartObjectsNonMatchingPodImage},
		{"MatchingPodImages", testArg1, expectedRestartObjectsMatchingPodImage},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := ReconcilerWrapper{}
			got := re.FilterNameSpaceNeedAttention(tt.args.clientset, tt.args.namespaces, tt.args.config, tt.args.k8sapi)

			if !IsRestartObjectSame(got, tt.want) {
				t.Errorf("FilterNameSpaceNeedAttention() = %v, want %v", got, tt.want)
			}
		})
	}
}
