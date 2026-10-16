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

// StaggerSleep is the sleep used by staggered mode for the settle delay before
// the first status poll and for the waits between polls. It is a variable so
// tests can replace it.
var StaggerSleep = time.Sleep

// RestartObjectGated restarts a single object and then waits for it to be
// strictly healthy (see IsHealthyStrict) or exhaust retries before returning.
// The wait is forced regardless of config.CheckStatus; MaxRetry/RetryDuration fall back to defaults when unset.
// A settle delay of one RetryDuration precedes the first status poll because
// Argo Rollouts can still report Healthy right after the restart annotation is
// patched. If the restart command itself fails the wait is skipped.
// This is the per-object primitive behind FLIPPY_STAGGER_NAMESPACES.
func RestartObjectGated(k8s k8s.K8sAPI, processor RestartProcessorInterface, config crdv1.StatusCheckConfig, namespace string, objectName string) {
	restartConfig := config
	restartConfig.CheckStatus = false // the forced wait below gates instead
	waitConfig := staggerWaitConfig(config)

	if err := processor.RestartObject(k8s, restartConfig, namespace, objectName, 0); err != nil {
		log.WithFields(log.Fields{common.NAMESPACE: namespace, common.NAME: objectName}).Warn("Staggered restart: restart command failed, skipping health wait for this object")
		return
	}
	StaggerSleep(time.Duration(waitConfig.RetryDuration) * time.Second)
	processor.WaitForHealthy(k8s, waitConfig, namespace, objectName)
}

// IsHealthyStrict reports whether a rollout status output means the rollout has
// fully completed: a Deployment prints "successfully rolled out" and an Argo
// Rollout prints exactly "Healthy". Unlike IsRestartGood it does not accept
// in-progress outputs such as "1 old replicas are pending termination" or
// "0 of 1 updated replicas are available", so the staggered gate really waits
// for the object it just restarted. IsRestartGood is unchanged and remains the
// check for the default mode, Preconditions and PostFilterRestarts.
func IsHealthyStrict(output string) bool {
	output = strings.TrimSpace(output)
	return strings.Contains(output, "successfully rolled out") || output == "Healthy"
}

// waitForHealthy polls status until IsHealthyStrict passes or MaxRetry retries
// are exhausted, sleeping RetryDuration between polls. Staggered mode only.
func waitForHealthy(status func() (string, error), waitConfig crdv1.StatusCheckConfig, logFields log.Fields) {
	for retryCount := 0; ; retryCount++ {
		output, err := status()
		if err != nil {
			log.WithFields(logFields).WithField(common.RETRY_COUNT, retryCount).Error("Failed to fetch status", err)
		}
		if IsHealthyStrict(output) {
			log.WithFields(logFields).Info("Restart completed")
			return
		}
		if retryCount >= waitConfig.MaxRetry {
			log.WithFields(logFields).WithField(common.RETRY_COUNT, retryCount).Info("Restart retry timed out. Status - " + strings.TrimSpace(output))
			return
		}
		log.WithFields(logFields).Infof("Not healthy yet, retrying in %ds. Status - %s", waitConfig.RetryDuration, strings.TrimSpace(output))
		StaggerSleep(time.Duration(waitConfig.RetryDuration) * time.Second)
	}
}

func (RestartDeploymentWrapper) WaitForHealthy(k8s k8s.K8sAPI, waitConfig crdv1.StatusCheckConfig, namespace string, restartObjectName string) {
	waitForHealthy(func() (string, error) {
		return k8s.RolloutDeploymentStatus(common.KubeconfigPath, namespace, restartObjectName)
	}, waitConfig, log.Fields{common.TYPE: common.DEPLOYMENT, common.NAME: restartObjectName, common.NAMESPACE: namespace})
}

func (RestartRolloutWrapper) WaitForHealthy(k8s k8s.K8sAPI, waitConfig crdv1.StatusCheckConfig, namespace string, restartObjectName string) {
	waitForHealthy(func() (string, error) {
		return k8s.RolloutArogRolloutStatus(common.KubeconfigPath, namespace, restartObjectName)
	}, waitConfig, log.Fields{common.TYPE: common.ARGO_ROLLOUT, common.NAME: restartObjectName, common.NAMESPACE: namespace})
}

// staggerWaitConfig forces a status check, falling back to defaults when the
// CRD config does not specify retry parameters.
func staggerWaitConfig(config crdv1.StatusCheckConfig) crdv1.StatusCheckConfig {
	wait := crdv1.StatusCheckConfig{
		CheckStatus:   true,
		MaxRetry:      config.MaxRetry,
		RetryDuration: config.RetryDuration,
	}
	if wait.MaxRetry <= 0 {
		wait.MaxRetry = common.DefaultStaggerMaxRetry
	}
	if wait.RetryDuration <= 0 {
		wait.RetryDuration = common.DefaultStaggerRetryDurationSeconds
	}
	return wait
}

func (RestartDeploymentWrapper) RestartObject(k8s k8s.K8sAPI, restartConfig crdv1.StatusCheckConfig, namespace string, restartObjectName string, retryCount int) error {
	log.Infof("Restarting deployment %s in namespace %s", restartObjectName, namespace)
	output, err := k8s.RolloutRestartDeployment(common.KubeconfigPath, namespace, restartObjectName)
	if err != nil {
		log.Errorf("Failed to restart deployment %s in namespace %s. Error - %s", restartObjectName, namespace, err)
	}
	log.Info(output)
	if restartConfig.CheckStatus {
		RestartDeploymentProcessor.WaitForRestartToBeComplete(k8s, restartConfig, namespace, restartObjectName, retryCount)
	}
	return err
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

func (RestartRolloutWrapper) RestartObject(k8s k8s.K8sAPI, restartConfig crdv1.StatusCheckConfig, namespace string, restartObjectName string, retryCount int) error {
	log.Infof("Restarting rollout %s in namespace %s", restartObjectName, namespace)
	output, err := k8s.RolloutRestartArgoRollouts(common.KubeconfigPath, namespace, restartObjectName)
	if err != nil {
		log.Errorf("Failed to restart rollout %s in namespace %s. Error - %s", restartObjectName, namespace, err)
	}
	log.Info(output)

	if restartConfig.CheckStatus {
		RestartRolloutProcessor.WaitForRestartToBeComplete(k8s, restartConfig, namespace, restartObjectName, retryCount)
	}
	return err
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
			RestartRolloutProcessor.WaitForRestartToBeComplete(k8s, restartConfig, namespace, restartObjectName, retryCount+1)
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
