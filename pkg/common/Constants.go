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

// StaggerMaxParallelGroupsEnvVar caps how many asset-alias groups rotate concurrently in staggered mode.
const StaggerMaxParallelGroupsEnvVar = "FLIPPY_STAGGER_MAX_PARALLEL_GROUPS"
const DefaultStaggerMaxParallelGroups = 5

// AssetAliasAnnotation is the Namespace annotation that identifies the owning asset.
// Namespaces sharing a value are rotated sequentially as one group in staggered mode.
const AssetAliasAnnotation = "iks.intuit.com/service-asset-alias"

// StaggerNamespacesEnabled, when true, groups namespaces by asset alias and
// rotates each group one namespace at a time (Deployments then Rollouts, one
// object at a time with a health wait). Groups run in parallel up to
// StaggerMaxParallelGroups. Read once at startup from FLIPPY_STAGGER_NAMESPACES.
var StaggerNamespacesEnabled = ParseStaggerNamespacesEnabled(os.Getenv(StaggerNamespacesEnvVar))

// StaggerMaxParallelGroups is the concurrency cap for alias groups in staggered mode.
var StaggerMaxParallelGroups = ParseStaggerMaxParallelGroups(os.Getenv(StaggerMaxParallelGroupsEnvVar))

// ParseStaggerMaxParallelGroups parses the env var value. Missing, invalid or
// non-positive values yield DefaultStaggerMaxParallelGroups.
func ParseStaggerMaxParallelGroups(value string) int {
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		return DefaultStaggerMaxParallelGroups
	}
	return n
}

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
	// Healthy is true when the objects were reporting healthy before Flippy
	// touched them. Only healthy objects get the forced health wait in
	// staggered mode; already-unhealthy ones are restarted without waiting.
	Healthy bool
}

var IgnoreMetadataKey string
