package RestartProcessor

import (
	crdv1 "keikoproj.intuit.com/Flippy/api/v1"
	"keikoproj.intuit.com/Flippy/pkg/common"
	"keikoproj.intuit.com/Flippy/pkg/k8s-utils/k8s"
)

type RestartProcessorInterface interface {
	RestartObject(k8s k8s.K8sAPI, restartConfig crdv1.StatusCheckConfig, namespace string, restartObjectName string, retryCount int)
	WaitForRestartToBeComplete(k8s k8s.K8sAPI, restartConfig crdv1.StatusCheckConfig, namespace string, restartObjectName string, retryCount int)
	Restart(k8s k8s.K8sAPI, restarts common.RestartObjects)
}

type RestartDeploymentWrapper struct{}

type RestartRolloutWrapper struct{}

var RestartDeploymentProcessor RestartProcessorInterface = RestartDeploymentWrapper{}

var RestartRolloutProcessor RestartProcessorInterface = RestartRolloutWrapper{}
