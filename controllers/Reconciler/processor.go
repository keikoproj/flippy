package Reconciler

import (
	crdv1 "github.com/keikoproj/flippy/api/v1"
	"github.com/keikoproj/flippy/controllers/RestartProcessor"
	"github.com/keikoproj/flippy/pkg/common"
	"github.com/keikoproj/flippy/pkg/k8s-utils/k8s"
	log "github.com/sirupsen/logrus"
	"strings"
)

func (ReconcilerWrapper) ProcessRestart(k8s k8s.K8sAPI, clientset k8s.ClientSet, config crdv1.FlippyConfig) error {

	//Get all namespaces which are mesh enabled
	meshEnabledNamespaces, err := k8s.GetNamespaceWithLabelFilter(clientset.K8sClientSet, config.Spec.ProcessFilter.NamespaceLabels)

	if err != nil {
		log.Debug("Failed to get Namespaces filtered with label", err)
		return err
	}

	restarts := Process.FilterNameSpaceNeedAttention(clientset, meshEnabledNamespaces, config, k8s)

	var preRestarts []common.RestartObjects

	//Make sure restart map is filled. If restart map is empty then restart is not required.
	isRestartRequired := false
	for _, restart := range restarts {
		if len(restart.NamespaceObjects) > 0 {
			isRestartRequired = true
		}
	}

	if isRestartRequired {
		for _, postFilterRestart := range config.Spec.PostFilterRestarts {
			preRestartDeployemt := make(map[string][]string)
			preRestartDeployemt[postFilterRestart.K8S.Namespace] = append(preRestartDeployemt[postFilterRestart.K8S.Namespace], postFilterRestart.K8S.Name)

			preRestarts = append(preRestarts, common.RestartObjects{
				Type:             postFilterRestart.K8S.Type,
				NamespaceObjects: preRestartDeployemt,
				RestartConfig:    postFilterRestart.StatusCheckConfig,
			})
		}

		for _, preRestart := range preRestarts {
			switch strings.ToLower(preRestart.Type) {
			case common.DEPLOYMENT:
				RestartProcessor.RestartDeploymentProcessor.Restart(k8s, preRestart)
			case common.ARGO_ROLLOUT:
				RestartProcessor.RestartRolloutProcessor.Restart(k8s, preRestart)
			default:
				log.Error("Failed to process " + preRestart.Type)
			}
		}

		for _, restart := range restarts {
			switch strings.ToLower(restart.Type) {
			case common.DEPLOYMENT:
				RestartProcessor.RestartDeploymentProcessor.Restart(k8s, restart)
			case common.ARGO_ROLLOUT:
				RestartProcessor.RestartRolloutProcessor.Restart(k8s, restart)
			default:
				log.Error("Failed to process " + restart.Type)
			}
		}

		log.Info("PreRestart process - ", preRestarts)
		log.Info("Restart Map - ", restarts)
	} else {
		log.Info("Restart is skipped because restart map is empty")
		return nil
	}
	return nil
}

func (ReconcilerWrapper) ProcessNamespaceRestarts(k8s k8s.K8sAPI, restartObjects []common.RestartObjects) {

	var restartProcessor RestartProcessor.RestartProcessorInterface
	for _, restartObject := range restartObjects {
		switch restartObject.Type {
		case common.DEPLOYMENT:
			restartProcessor = RestartProcessor.RestartDeploymentProcessor
		case common.ARGO_ROLLOUT:
			restartProcessor = RestartProcessor.RestartRolloutProcessor
		default:
			log.Infof("Provide object type %s is not currently supported", restartObject.Type)
		}

		for namespace, restartObjects := range restartObject.NamespaceObjects {
			for _, restartObjectName := range restartObjects {
				restartProcessor.RestartObject(k8s, restartObject.RestartConfig, namespace, restartObjectName, 0)
			}
		}
	}
}
