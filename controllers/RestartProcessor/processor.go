package RestartProcessor

import (
	crdv1 "github.com/keikoproj/flippy/api/v1"
	"github.com/keikoproj/flippy/pkg/common"
	"github.com/keikoproj/flippy/pkg/k8s-utils/k8s"
	log "github.com/sirupsen/logrus"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (RestartDeploymentWrapper) Restart(k8s k8s.K8sAPI, restart common.RestartObjects) {

	if strings.ToLower(restart.Type) == common.DEPLOYMENT {
		if common.StaggerNamespacesEnabled {
			restartNamespacesStaggered(k8s, RestartDeploymentProcessor, common.DEPLOYMENT, restart)
			return
		}
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
		if common.StaggerNamespacesEnabled {
			restartNamespacesStaggered(k8s, RestartRolloutProcessor, common.ARGO_ROLLOUT, restart)
			return
		}
		for namespace, objects := range restart.NamespaceObjects {
			for _, objectName := range objects {
				RestartRolloutProcessor.RestartObject(k8s, restart.RestartConfig, namespace, objectName, 0)
			}
		}
	} else {
		log.Error("Found " + restart.Type + " while processing " + common.ARGO_ROLLOUT)
	}
}

// staggerSleep is the sleep used for the settle delay before the first status
// poll in staggered mode. It is a variable so tests can replace it.
var staggerSleep = time.Sleep

// restartNamespacesStaggered processes namespaces one at a time in sorted order.
// Within a namespace each object is restarted and then forced through
// WaitForRestartToBeComplete (regardless of the CRD's CheckStatus) before the
// next object is touched; the next namespace starts only after every object in
// the current one is healthy or has exhausted its retries.
// Used when FLIPPY_STAGGER_NAMESPACES=true.
func restartNamespacesStaggered(k8s k8s.K8sAPI, processor RestartProcessorInterface, objectType string, restart common.RestartObjects) {
	namespaces := make([]string, 0, len(restart.NamespaceObjects))
	for namespace := range restart.NamespaceObjects {
		namespaces = append(namespaces, namespace)
	}
	sort.Strings(namespaces)

	// Restart without the per-object wait inside RestartObject; the forced wait below gates instead.
	restartConfig := restart.RestartConfig
	restartConfig.CheckStatus = false

	waitConfig := staggerWaitConfig(restart.RestartConfig)

	log.WithFields(log.Fields{common.TYPE: objectType}).Infof("Namespace staggering enabled. Processing %d namespace(s) in order: %v", len(namespaces), namespaces)

	for _, namespace := range namespaces {
		objects := restart.NamespaceObjects[namespace]
		logFields := log.Fields{common.TYPE: objectType, common.NAMESPACE: namespace}
		log.WithFields(logFields).Infof("Staggered restart: starting namespace with %d object(s)", len(objects))

		for _, objectName := range objects {
			if err := processor.RestartObject(k8s, restartConfig, namespace, objectName, 0); err != nil {
				// The restart command itself failed (object gone, RBAC, etc.). Waiting on its
				// status would only burn the retry budget, so move on.
				log.WithFields(logFields).WithField(common.NAME, objectName).Warn("Staggered restart: restart command failed, skipping health wait for this object")
				continue
			}
			// Give the controller a chance to observe the restart before trusting the
			// first status read. Argo Rollouts in particular can still report Healthy
			// for a moment after the restart annotation is patched.
			staggerSleep(time.Duration(waitConfig.RetryDuration) * time.Second)
			processor.WaitForRestartToBeComplete(k8s, waitConfig, namespace, objectName, 0)
		}

		log.WithFields(logFields).Info("Staggered restart: namespace complete")
	}
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
