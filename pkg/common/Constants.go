package common

import (
	"os"
	"strconv"

	crdv1 "github.com/keikoproj/flippy/api/v1"
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

// StaggerNamespacesEnvVar is the cluster-wide opt-in for namespace-level staggering of restarts.
const StaggerNamespacesEnvVar = "FLIPPY_STAGGER_NAMESPACES"

// Defaults used for the forced per-namespace health wait when the CRD's
// StatusCheckConfig does not provide MaxRetry / RetryDuration.
const DefaultStaggerMaxRetry = 10
const DefaultStaggerRetryDurationSeconds = 30

// StaggerNamespacesEnabled, when true, makes the restart processors handle one
// namespace at a time in sorted order, waiting for every object in a namespace
// to be healthy before moving on. Read once at startup from FLIPPY_STAGGER_NAMESPACES.
var StaggerNamespacesEnabled = ParseStaggerNamespacesEnabled(os.Getenv(StaggerNamespacesEnvVar))

// ParseStaggerNamespacesEnabled parses the env var value. Missing or invalid values yield false.
func ParseStaggerNamespacesEnabled(value string) bool {
	enabled, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}
	return enabled
}

type RestartObjects struct {
	Type string
	//Map of Namespace and list of objects to restart
	NamespaceObjects map[string][]string
	RestartConfig    crdv1.StatusCheckConfig
}

var IgnoreMetadataKey string
