package Reconciler

import (
	"errors"
	crdv1 "github.com/keikoproj/flippy/api/v1"
	"github.com/keikoproj/flippy/controllers/RestartProcessor"
	"github.com/keikoproj/flippy/pkg/common"
	"github.com/keikoproj/flippy/pkg/k8s-utils/k8s"
	"github.com/keikoproj/flippy/pkg/k8s-utils/utils"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"strings"
	"time"
)

var dockerImages []string

func (ReconcilerWrapper) Handle(req crdv1.FlippyConfig, clientsetWrapper k8s.ClientSet, k8sUtil k8s.K8sAPI) error {
	if Process.IsProxyVersionChange(req.Spec.ImageList) {
		log.Info("ImageList Drift Found.")
		timeStart := time.Now()
		//is pre-condition verified
		for _, condition := range req.Spec.Preconditions {
			if !Process.IsPreConditionSatisfied(k8s.K8s, clientsetWrapper.K8sClientSet, condition, 0) {
				log.WithFields(log.Fields{
					"Precondition": condition,
				}).Error("Precondition failed.", condition)
				err := errors.New("precondition failed")
				return err
			}
		}
		err := Process.ProcessRestart(k8sUtil, clientsetWrapper, req)
		if err != nil {
			log.Error(err)
			return err
		}
		log.Info("Restart Process Completed in ", time.Now().Sub(timeStart), "sec")

	} else {
		log.Info("No ImageList Drift found.")
	}
	return nil
}

func (ReconcilerWrapper) IsProxyVersionChange(imageList []string) bool {

	if len(imageList) == 0 {
		log.Info("Empty Image List")
		return false
	}

	if len(dockerImages) == 0 {
		dockerImages = imageList
		return true
	} else if len(dockerImages) != len(imageList) {
		dockerImages = imageList
		return true
	} else {
		foundCount := 0
		for _, image := range imageList {
			if utils.StringArrayContains(dockerImages, image) {
				foundCount++
			}
		}

		if foundCount == (len(imageList)) {
			log.Info("Desire container list matched. Current - ", dockerImages, " Desire - ", imageList)
			return false
		} else {
			return true
		}
	}
}

func (ReconcilerWrapper) IsPreConditionSatisfied(k8sapi k8s.K8sAPI, clientset kubernetes.Interface, condition crdv1.FlippyCondition, retryCount int) bool {
	//Make sure pre-condition satisfied
	if retryCount < condition.StatusCheckConfig.MaxRetry {

		switch strings.ToLower(condition.K8S.Type) {
		case common.DEPLOYMENT:
			RestartProcessor.RestartDeploymentProcessor.WaitForRestartToBeComplete(k8sapi, condition.StatusCheckConfig, condition.K8S.Namespace, condition.K8S.Name, 0)

			deployment, err := k8sapi.GetDeployment(clientset, condition.K8S.Namespace, condition.K8S.Name)

			if err != nil {
				log.Error("Failed to get "+condition.K8S.Name+" status", err)
				return false
			}

			if deployment.Status.UnavailableReplicas != 0 && deployment.Status.Replicas-deployment.Status.AvailableReplicas != 0 {
				retryCount = retryCount + 1
				log.Info("Retrying PreCondition Check. Count - ", retryCount)
				time.Sleep(time.Duration(condition.StatusCheckConfig.RetryDuration) * time.Second)
				return Process.IsPreConditionSatisfied(k8sapi, clientset, condition, retryCount)
			}
			return true
		default:
			log.Error("Skipping. "+condition.K8S.Type+" is not supported.", condition)
			return true
		}
	} else if !condition.StatusCheckConfig.CheckStatus {
		return true
	}
	return false
}
