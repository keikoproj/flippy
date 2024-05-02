package common

import (
	crdv1 "github.com/keikoproj/flippy/api/v1"
	"os"
)

var KubeconfigPath = os.Getenv("KUBECONFIG")

const ARGO_ROLLOUT = "argorollout"
const DEPLOYMENT = "deployment"
const TYPE = "Type"
const NAMESPACE = "Namespace"
const NAME = "Name"
const WAIT_FOR_RESTART_TO_COMPLETE = "WaitForRestartToComplete"
const MAX_RETRY_COUNT = "MaxRetryCount"
const RETRY_DURATION = "RetryDuration"
const RETRY_COUNT = "RetryCount"

type RestartObjects struct {
	Type string
	//Map of Namespace and list of objects to restart
	NamespaceObjects map[string][]string
	RestartConfig    crdv1.StatusCheckConfig
}

var IgnoreMetadata string
