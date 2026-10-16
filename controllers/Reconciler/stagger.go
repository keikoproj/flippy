package Reconciler

import (
	"sort"
	"strings"
	"sync"

	crdv1 "github.com/keikoproj/flippy/api/v1"
	"github.com/keikoproj/flippy/controllers/RestartProcessor"
	"github.com/keikoproj/flippy/pkg/common"
	"github.com/keikoproj/flippy/pkg/k8s-utils/k8s"
	log "github.com/sirupsen/logrus"
)

// staggerObject is one Deployment or Rollout to restart within a namespace.
type staggerObject struct {
	name   string
	config crdv1.StatusCheckConfig
	// gated objects are restarted then waited on; ungated (already unhealthy
	// before Flippy touched them) are restarted without waiting, as on master.
	gated bool
}

// staggerNamespace holds everything to restart in one namespace, in the order
// it will be processed: all Deployments, then all Rollouts.
type staggerNamespace struct {
	name        string
	deployments []staggerObject
	rollouts    []staggerObject
}

// staggerGroup is a set of namespaces sharing an asset alias. Namespaces in a
// group are processed one at a time, in sorted order.
type staggerGroup struct {
	key        string
	namespaces []*staggerNamespace
}

func (ReconcilerWrapper) ProcessRestartsStaggered(k8sapi k8s.K8sAPI, clientset k8s.ClientSet, restarts []common.RestartObjects) {
	byNamespace := collectStaggerNamespaces(restarts)
	if len(byNamespace) == 0 {
		log.Info("Staggered restart: nothing to restart")
		return
	}

	aliases := namespaceAliases(k8sapi, clientset)
	groups := groupByAlias(byNamespace, aliases)

	maxParallel := common.StaggerMaxParallelGroups
	if maxParallel <= 0 {
		maxParallel = common.DefaultStaggerMaxParallelGroups
	}

	log.Infof("Staggered restart: %d namespace(s) in %d group(s), up to %d group(s) in parallel", len(byNamespace), len(groups), maxParallel)
	for _, g := range groups {
		names := make([]string, 0, len(g.namespaces))
		for _, ns := range g.namespaces {
			names = append(names, ns.name)
		}
		log.Infof("Staggered restart: group %q -> %v", g.key, names)
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxParallel)
	for _, g := range groups {
		wg.Add(1)
		go func(g staggerGroup) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			processStaggerGroup(k8sapi, g)
		}(g)
	}
	wg.Wait()
	log.Info("Staggered restart: all groups complete")
}

// collectStaggerNamespaces folds the four per-type/per-health restart lists into
// one entry per namespace, Deployments before Rollouts, healthy before unhealthy.
func collectStaggerNamespaces(restarts []common.RestartObjects) map[string]*staggerNamespace {
	byNamespace := make(map[string]*staggerNamespace)
	get := func(name string) *staggerNamespace {
		if ns, ok := byNamespace[name]; ok {
			return ns
		}
		ns := &staggerNamespace{name: name}
		byNamespace[name] = ns
		return ns
	}

	// Two passes so healthy objects precede unhealthy ones within each type.
	for _, wantHealthy := range []bool{true, false} {
		for _, restart := range restarts {
			if restart.Healthy != wantHealthy {
				continue
			}
			for namespace, names := range restart.NamespaceObjects {
				ns := get(namespace)
				for _, name := range names {
					obj := staggerObject{name: name, config: restart.RestartConfig, gated: restart.Healthy}
					switch strings.ToLower(restart.Type) {
					case common.DEPLOYMENT:
						ns.deployments = append(ns.deployments, obj)
					case common.ARGO_ROLLOUT:
						ns.rollouts = append(ns.rollouts, obj)
					default:
						log.Errorf("Staggered restart: unsupported type %q for %s/%s", restart.Type, namespace, name)
					}
				}
			}
		}
	}
	return byNamespace
}

// namespaceAliases returns namespace name -> asset alias annotation value.
// On lookup failure it returns an empty map so every namespace becomes its own group.
func namespaceAliases(k8sapi k8s.K8sAPI, clientset k8s.ClientSet) map[string]string {
	aliases := make(map[string]string)
	list, err := k8sapi.GetNamespaces(clientset.K8sClientSet)
	if err != nil || list == nil {
		log.Warn("Staggered restart: failed to list namespaces for asset alias grouping; treating each namespace as its own group. ", err)
		return aliases
	}
	for _, ns := range list.Items {
		if alias, ok := ns.Annotations[common.AssetAliasAnnotation]; ok && strings.TrimSpace(alias) != "" {
			aliases[ns.Name] = strings.TrimSpace(alias)
		}
	}
	return aliases
}

// groupByAlias buckets namespaces by alias. Namespaces without an alias get a
// unique key so each is its own group. Groups and their namespaces are sorted
// for deterministic ordering.
func groupByAlias(byNamespace map[string]*staggerNamespace, aliases map[string]string) []staggerGroup {
	grouped := make(map[string][]*staggerNamespace)
	for name, ns := range byNamespace {
		key := "namespace:" + name
		if alias, ok := aliases[name]; ok {
			key = "alias:" + alias
		}
		grouped[key] = append(grouped[key], ns)
	}

	groups := make([]staggerGroup, 0, len(grouped))
	for key, namespaces := range grouped {
		sort.Slice(namespaces, func(i, j int) bool { return namespaces[i].name < namespaces[j].name })
		groups = append(groups, staggerGroup{key: key, namespaces: namespaces})
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].key < groups[j].key })
	return groups
}

// processStaggerGroup rotates a group's namespaces one at a time.
func processStaggerGroup(k8sapi k8s.K8sAPI, group staggerGroup) {
	for _, ns := range group.namespaces {
		fields := log.Fields{"Group": group.key, common.NAMESPACE: ns.name}
		log.WithFields(fields).Infof("Staggered restart: starting namespace (%d deployment(s), %d rollout(s))", len(ns.deployments), len(ns.rollouts))
		processStaggerObjects(k8sapi, RestartProcessor.RestartDeploymentProcessor, ns.name, ns.deployments)
		processStaggerObjects(k8sapi, RestartProcessor.RestartRolloutProcessor, ns.name, ns.rollouts)
		log.WithFields(fields).Info("Staggered restart: namespace complete")
	}
}

// processStaggerObjects restarts objects one at a time. Gated objects wait for
// health before the next one starts; ungated ones are restarted without waiting.
func processStaggerObjects(k8sapi k8s.K8sAPI, processor RestartProcessor.RestartProcessorInterface, namespace string, objects []staggerObject) {
	for _, obj := range objects {
		if obj.gated {
			RestartProcessor.Gated(k8sapi, processor, obj.config, namespace, obj.name)
			continue
		}
		cfg := obj.config
		cfg.CheckStatus = false
		_ = processor.RestartObject(k8sapi, cfg, namespace, obj.name, 0)
	}
}
