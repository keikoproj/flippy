package RestartProcessor

import (
	crdv1 "github.com/keikoproj/flippy/api/v1"
	"github.com/keikoproj/flippy/pkg/common"
	"github.com/keikoproj/flippy/pkg/k8s-utils/k8s"
)

type RestartProcessorInterface interface {
	// RestartObject issues the rollout restart and, when restartConfig.CheckStatus is set,
	// waits for it to complete. It returns the error from the restart command itself.
	RestartObject(k8s k8s.K8sAPI, restartConfig crdv1.StatusCheckConfig, namespace string, restartObjectName string, retryCount int) error
	WaitForRestartToBeComplete(k8s k8s.K8sAPI, restartConfig crdv1.StatusCheckConfig, namespace string, restartObjectName string, retryCount int)
	Restart(k8s k8s.K8sAPI, restarts common.RestartObjects)
}

// Gated is the restart-then-wait primitive used by staggered mode. It is a
// variable so the reconciler tests can stub it.
var Gated = RestartObjectGated

type RestartDeploymentWrapper struct{}

type RestartRolloutWrapper struct{}

var RestartDeploymentProcessor RestartProcessorInterface = RestartDeploymentWrapper{}

var RestartRolloutProcessor RestartProcessorInterface = RestartRolloutWrapper{}
