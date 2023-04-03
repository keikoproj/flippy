package RestartProcessor

import (
	crdv1 "github.com/keikoproj/flippy/api/v1"
	"github.com/keikoproj/flippy/pkg/common"
	"github.com/keikoproj/flippy/pkg/k8s-utils/k8s"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

func (RestartDeploymentWrapper) Restart(k8s k8s.K8sAPI, restart common.RestartObjects) {

	if strings.ToLower(restart.Type) == common.DEPLOYMENT {
		for namespace, objects := range restart.NamespaceObjects {
			for _, objectName := range objects {
				RestartDeploymentProcessor.RestartObject(k8s, restart.RestartConfig, namespace, objectName, 0)
			}
		}
	} else {
		log.Error("Found " + restart.Type + " while processing " + common.DEPLOYMENT)
	}

}

func (RestartRolloutWrapper) Restart(k8s k8s.K8sAPI, restart common.RestartObjects) {

	if strings.ToLower(restart.Type) == common.ARGO_ROLLOUT {
		for namespace, objects := range restart.NamespaceObjects {
			for _, objectName := range objects {
				RestartRolloutProcessor.RestartObject(k8s, restart.RestartConfig, namespace, objectName, 0)
			}
		}
	} else {
		log.Error("Found " + restart.Type + " while processing " + common.ARGO_ROLLOUT)
	}
}

func (RestartDeploymentWrapper) RestartObject(k8s k8s.K8sAPI, restartConfig crdv1.StatusCheckConfig, namespace string, restartObjectName string, retryCount int) {
	log.Infof("Restarting deployment %s in namespace %s", restartObjectName, namespace)
	output, err := k8s.RolloutRestartDeployment(common.KubeconfigPath, namespace, restartObjectName)
	if err != nil {
		log.Errorf("Failed to restart deployment %s in namespace %s. Error - %s", restartObjectName, namespace, err)
	}
	log.Info(output)
	if restartConfig.CheckStatus {
		RestartDeploymentProcessor.WaitForRestartToBeComplete(k8s, restartConfig, namespace, restartObjectName, retryCount)
	}
}

func (RestartDeploymentWrapper) WaitForRestartToBeComplete(k8s k8s.K8sAPI, restartConfig crdv1.StatusCheckConfig, namespace string, restartObjectName string, retryCount int) {
	output, err := k8s.RolloutDeploymentStatus(common.KubeconfigPath, namespace, restartObjectName)

	logFields := log.Fields{
		common.TYPE:      common.DEPLOYMENT,
		common.NAME:      restartObjectName,
		common.NAMESPACE: namespace,
	}

	if err != nil {
		logFields[common.RETRY_COUNT] = retryCount
		log.WithFields(logFields).Error("Failed to fetch status", err)
	}

	if !IsRestartGood(output) {
		if retryCount < restartConfig.MaxRetry {
			log.WithFields(logFields).Info("Retrying restart")
			time.Sleep(time.Duration(restartConfig.RetryDuration) * time.Second)
			RestartDeploymentProcessor.WaitForRestartToBeComplete(k8s, restartConfig, namespace, restartObjectName, retryCount+1)
		} else {
			logFields[common.RETRY_COUNT] = retryCount
			log.WithFields(logFields).Info("Restart retry timed out")
		}
		log.Info(output, err)
	} else {
		log.WithFields(logFields).Info("Restart completed")
	}
}

func (RestartRolloutWrapper) RestartObject(k8s k8s.K8sAPI, restartConfig crdv1.StatusCheckConfig, namespace string, restartObjectName string, retryCount int) {
	log.Infof("Restarting rollout %s in namespace %s", restartObjectName, namespace)
	output, err := k8s.RolloutRestartArgoRollouts(common.KubeconfigPath, namespace, restartObjectName)
	if err != nil {
		log.Errorf("Failed to restart rollout %s in namespace %s. Error - %s", restartObjectName, namespace, err)
	}
	log.Info(output)

	if restartConfig.CheckStatus {
		RestartRolloutProcessor.WaitForRestartToBeComplete(k8s, restartConfig, namespace, restartObjectName, retryCount)
	}
}

func (RestartRolloutWrapper) WaitForRestartToBeComplete(k8s k8s.K8sAPI, restartConfig crdv1.StatusCheckConfig, namespace string, restartObjectName string, retryCount int) {
	output, err := k8s.RolloutArogRolloutStatus(common.KubeconfigPath, namespace, restartObjectName)

	logFields := log.Fields{
		common.TYPE:      common.ARGO_ROLLOUT,
		common.NAME:      restartObjectName,
		common.NAMESPACE: namespace,
	}

	if err != nil {
		logFields[common.RETRY_COUNT] = retryCount
		log.WithFields(logFields).Error("Failed to fetch status", err)
	}

	if !IsRestartGood(output) {
		if retryCount < restartConfig.MaxRetry {
			log.WithFields(logFields).Info("Retrying restart")
			time.Sleep(time.Duration(restartConfig.RetryDuration) * time.Second)
			RestartRolloutProcessor.WaitForRestartToBeComplete(k8s, restartConfig, namespace, restartObjectName, retryCount)
		} else {
			logFields[common.RETRY_COUNT] = retryCount
			log.WithFields(logFields).Info("Restart retry timed out")
		}
		log.WithFields(logFields).Info("Status - "+output+" Error - ", err)
	} else {
		log.WithFields(logFields).Info("Restart completed")
	}
}

func IsRestartGood(output string) bool {
	log.Debugf("Restart status output %s", output)
	output = strings.TrimSpace(output)
	if output == "" {
		return false
	} else if strings.Contains(output, "successfully rolled out") || strings.Contains(output, "updated replicas are available") || output == "Healthy" {
		return true
	} else {
		for _, outputLine := range strings.Split(output, "\n") {
			if strings.Contains(outputLine, "rollout to finish:") {
				activePod, err := strconv.Atoi(strings.Split(strings.TrimSpace(strings.Split(outputLine, ":")[1]), " ")[0])
				if err == nil {
					if activePod > 0 {
						return true
					}
				}
			}
		}
	}
	return false
}
