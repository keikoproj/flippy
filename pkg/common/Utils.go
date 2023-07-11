package common

import (
	log "github.com/sirupsen/logrus"
	"time"
)

func GetReconcileDuration(flagReconcilerTime string) time.Duration {
	reconcileDuration, err := time.ParseDuration(flagReconcilerTime)
	if err != nil {
		log.Info("Failed to parse reconcile time. Setting up 10 hrs")
		reconcileDuration = 10 * time.Hour
	} else {
		log.Info("Setting Reconcile Time to " + reconcileDuration.String())
	}
	return reconcileDuration
}
