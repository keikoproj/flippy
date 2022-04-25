package Reconciler

import (
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	crdv1 "keikoproj.intuit.com/Flippy/api/v1"
	"keikoproj.intuit.com/Flippy/pkg/common"
	"keikoproj.intuit.com/Flippy/pkg/k8s-utils/k8s"
	"keikoproj.intuit.com/Flippy/pkg/k8s-utils/utils"
	"strings"
)

func (ReconcilerWrapper) FilterNameSpaceNeedAttention(clientset k8s.ClientSet, namespaces []string, config crdv1.FlippyConfig, k8sapi k8s.K8sAPI) []common.RestartObjects {

	healthyMapNamespaceToDeployments := make(map[string][]string)
	nonhealthyMapNamespaceToDeployments := make(map[string][]string)
	healthyMapNamespaceToArgoRollout := make(map[string][]string)
	nonhealthyMapNamespaceToArgoRollout := make(map[string][]string)

	for _, namespace := range namespaces {
		for _, objectType := range config.Spec.RestartObjects {
			if strings.ToLower(objectType.Type) == common.ARGO_ROLLOUT {
				rollouts, err := k8sapi.GetArgoRolloutsWithSpecAnnotationFilter(clientset.ArgoRolloutClientSet, namespace, config.Spec.ProcessFilter.Annotations)

				if err != nil {
					log.WithFields(log.Fields{common.TYPE: common.ARGO_ROLLOUT,
						common.NAMESPACE: namespace}).Error("Fail to get", err)
				} else {

					for _, rollout := range rollouts {
						rolloutStatus, err := k8sapi.RolloutArogRolloutStatus(common.KubeconfigPath, namespace, rollout)
						if err != nil {
							log.WithFields(log.Fields{common.TYPE: common.ARGO_ROLLOUT,
								common.NAME: rollout}).Error("Failed to get status", err)
						}

						podList, err := k8sapi.GetAllPodsInNamespace(clientset.K8sClientSet, namespace)
						if err != nil {
							log.Error("Fail to get running pods", err)
							continue
						}

						//We rely on argo rollout status rather than pod
						if Process.IsAnyPodContainsContainers(podList, config.Spec.ProcessFilter.Containers) {
							if !Process.IsAnyPodRunningWithProvidedDockerImage(podList, config.Spec.ImageList, config.Spec.ProcessFilter.Containers) {
								if strings.Contains(rolloutStatus, "Healthy") {
									healthyMapNamespaceToArgoRollout[namespace] = append(healthyMapNamespaceToArgoRollout[namespace], rollout)
								} else {
									nonhealthyMapNamespaceToArgoRollout[namespace] = append(nonhealthyMapNamespaceToArgoRollout[namespace], rollout)
								}
							}
						}
					}
				}
			}

			if strings.ToLower(objectType.Type) == common.DEPLOYMENT {
				deployments, err := k8sapi.GetDeploymentWithSpecAnnotationFilter(clientset.K8sClientSet, namespace, config.Spec.ProcessFilter.Annotations)
				if err != nil {
					log.WithFields(log.Fields{common.TYPE: common.DEPLOYMENT,
						common.NAMESPACE: namespace}).Error("Fail to get", err)

				} else {

					for _, deployment := range deployments {
						deploymentStatus, err := k8sapi.RolloutDeploymentStatus(common.KubeconfigPath, namespace, deployment)
						if err != nil {
							log.WithFields(log.Fields{common.TYPE: common.DEPLOYMENT,
								common.NAME: deployment}).Error("Failed to get status", err)
						}
						podList, err := k8sapi.GetAllPodsInNamespace(clientset.K8sClientSet, namespace)
						if err != nil {
							log.Error("Fail to get running pods", err)
							continue
						}

						//We rely on deployment status rather than pod
						if Process.IsAnyPodContainsContainers(podList, config.Spec.ProcessFilter.Containers) {
							if !Process.IsAnyPodRunningWithProvidedDockerImage(podList, config.Spec.ImageList, config.Spec.ProcessFilter.Containers) {
								if strings.Contains(deploymentStatus, "successfully rolled out") {
									healthyMapNamespaceToDeployments[namespace] = append(healthyMapNamespaceToDeployments[namespace], deployment)
								} else {
									nonhealthyMapNamespaceToDeployments[namespace] = append(nonhealthyMapNamespaceToDeployments[namespace], deployment)
								}
							}
						}
					}
				}
			}
		}

	}

	var deploymentRestartConfig, argoRestartConfig crdv1.StatusCheckConfig

	for _, restartObjects := range config.Spec.RestartObjects {
		switch strings.ToLower(restartObjects.Type) {
		case common.DEPLOYMENT:
			deploymentRestartConfig = restartObjects.StatusCheckConfig
		case common.ARGO_ROLLOUT:
			argoRestartConfig = restartObjects.StatusCheckConfig
		}
	}

	var restarts []common.RestartObjects

	restarts = append(restarts, common.RestartObjects{
		Type:             common.DEPLOYMENT,
		NamespaceObjects: healthyMapNamespaceToDeployments,
		RestartConfig:    deploymentRestartConfig,
	})

	restarts = append(restarts, common.RestartObjects{
		Type:             common.ARGO_ROLLOUT,
		NamespaceObjects: healthyMapNamespaceToArgoRollout,
		RestartConfig:    argoRestartConfig,
	})

	restarts = append(restarts, common.RestartObjects{
		Type:             common.DEPLOYMENT,
		NamespaceObjects: nonhealthyMapNamespaceToDeployments,
		RestartConfig: crdv1.StatusCheckConfig{
			CheckStatus:   false,
			MaxRetry:      0,
			RetryDuration: 0,
		},
	})

	restarts = append(restarts, common.RestartObjects{
		Type:             common.ARGO_ROLLOUT,
		NamespaceObjects: nonhealthyMapNamespaceToArgoRollout,
		RestartConfig: crdv1.StatusCheckConfig{
			CheckStatus:   false,
			MaxRetry:      0,
			RetryDuration: 0,
		},
	})

	return restarts
}

func (ReconcilerWrapper) IsAnyPodContainsContainers(podList []corev1.Pod, containers []string) bool {
	for _, pod := range podList {
		containersStatus := Process.IsPodContainContainers(pod, containers)
		match := 0
		for _, containerStatus := range containersStatus {
			if containerStatus.ContainerID != "" {
				match++
			}
		}
		if match > 0 && match == len(containersStatus) {
			return true
		}
	}
	return false
}

func (ReconcilerWrapper) IsPodContainContainers(pod corev1.Pod, containers []string) []corev1.ContainerStatus {
	var container []corev1.ContainerStatus
	for _, containerStatus := range pod.Status.ContainerStatuses {
		for _, containerName := range containers {
			if containerStatus.Name == containerName {
				container = append(container, containerStatus)
			}
		}
	}
	return container
}

func (ReconcilerWrapper) IsAnyPodRunningWithProvidedDockerImage(podList []corev1.Pod, dockerImages []string, containers []string) bool {

	for _, pod := range podList {
		containersStatus := Process.IsPodContainContainers(pod, containers)
		match := 0
		for _, containerStatus := range containersStatus {
			if containerStatus.ContainerID != "" && utils.StringArrayContains(dockerImages, containerStatus.Image) {
				match++
			}
		}
		if match > 0 {
			return true
		}
	}
	return false
}
