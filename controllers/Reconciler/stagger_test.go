package Reconciler

import (
	"errors"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	crdv1 "github.com/keikoproj/flippy/api/v1"
	"github.com/keikoproj/flippy/controllers/RestartProcessor"
	"github.com/keikoproj/flippy/pkg/common"
	"github.com/keikoproj/flippy/pkg/k8s-utils/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// aliasK8sFake serves namespace annotations for alias grouping on top of the shared fake.
type aliasK8sFake struct {
	K8sWrapperFakeSuccess
	aliases map[string]string // namespace -> alias ("" means no annotation)
	listErr error
}

func (f aliasK8sFake) GetNamespaces(kubernetes.Interface) (*corev1.NamespaceList, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	list := &corev1.NamespaceList{}
	for name, alias := range f.aliases {
		ns := corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}}
		if alias != "" {
			ns.Annotations = map[string]string{common.AssetAliasAnnotation: alias}
		}
		list.Items = append(list.Items, ns)
	}
	return list, nil
}

// recordingProcessor records every call, measures concurrency overall and per
// group, and flags any moment two namespaces of the same group overlap.
type recordingProcessor struct {
	kind        string
	shared      *recorder
	nsToGroup   map[string]string
	restartHold time.Duration
}

type recorder struct {
	mu              sync.Mutex
	sequence        []string
	active          int
	maxActive       int
	activeByGroup   map[string]int
	groupViolations []string
}

func newRecorder() *recorder { return &recorder{activeByGroup: map[string]int{}} }

func (p recordingProcessor) group(ns string) string {
	if g, ok := p.nsToGroup[ns]; ok {
		return g
	}
	return "namespace:" + ns
}

func (p recordingProcessor) RestartObject(_ k8s.K8sAPI, _ crdv1.StatusCheckConfig, ns, name string, _ int) error {
	r, g := p.shared, p.group(ns)
	r.mu.Lock()
	r.sequence = append(r.sequence, "restart-"+p.kind+":"+ns+"/"+name)
	r.active++
	if r.active > r.maxActive {
		r.maxActive = r.active
	}
	r.activeByGroup[g]++
	if r.activeByGroup[g] > 1 {
		r.groupViolations = append(r.groupViolations, g+" had "+ns+"/"+name+" concurrent with another namespace")
	}
	r.mu.Unlock()

	time.Sleep(p.restartHold)

	r.mu.Lock()
	r.active--
	r.activeByGroup[g]--
	r.mu.Unlock()
	if name == "gone" {
		return errors.New("not found")
	}
	return nil
}

func (p recordingProcessor) WaitForRestartToBeComplete(_ k8s.K8sAPI, _ crdv1.StatusCheckConfig, ns, name string, _ int) {
	p.shared.mu.Lock()
	defer p.shared.mu.Unlock()
	p.shared.sequence = append(p.shared.sequence, "wait-"+p.kind+":"+ns+"/"+name)
}

func (p recordingProcessor) Restart(k8s.K8sAPI, common.RestartObjects) {
	p.shared.mu.Lock()
	defer p.shared.mu.Unlock()
	p.shared.sequence = append(p.shared.sequence, "UNEXPECTED Restart() call")
}

// installRecorders swaps the package-level processors and settle sleep for the test.
func installRecorders(t *testing.T, nsToGroup map[string]string, hold time.Duration) *recorder {
	t.Helper()
	origDep, origRol, origSleep := RestartProcessor.RestartDeploymentProcessor, RestartProcessor.RestartRolloutProcessor, RestartProcessor.StaggerSleep
	origFlag, origCap := common.StaggerNamespacesEnabled, common.StaggerMaxParallelGroups
	rec := newRecorder()
	RestartProcessor.RestartDeploymentProcessor = recordingProcessor{kind: "dep", shared: rec, nsToGroup: nsToGroup, restartHold: hold}
	RestartProcessor.RestartRolloutProcessor = recordingProcessor{kind: "rol", shared: rec, nsToGroup: nsToGroup, restartHold: hold}
	RestartProcessor.StaggerSleep = func(time.Duration) {}
	common.StaggerNamespacesEnabled = true
	t.Cleanup(func() {
		RestartProcessor.RestartDeploymentProcessor, RestartProcessor.RestartRolloutProcessor, RestartProcessor.StaggerSleep = origDep, origRol, origSleep
		common.StaggerNamespacesEnabled, common.StaggerMaxParallelGroups = origFlag, origCap
	})
	return rec
}

func cfg(retry int) crdv1.StatusCheckConfig {
	return crdv1.StatusCheckConfig{CheckStatus: false, MaxRetry: retry, RetryDuration: 1}
}

// buildRestarts mirrors FilterNameSpaceNeedAttention's four-list shape.
func buildRestarts(healthyDep, healthyRol, unhealthyDep, unhealthyRol map[string][]string) []common.RestartObjects {
	return []common.RestartObjects{
		{Type: common.DEPLOYMENT, NamespaceObjects: healthyDep, RestartConfig: cfg(2), Healthy: true},
		{Type: common.ARGO_ROLLOUT, NamespaceObjects: healthyRol, RestartConfig: cfg(2), Healthy: true},
		{Type: common.DEPLOYMENT, NamespaceObjects: unhealthyDep, RestartConfig: cfg(0)},
		{Type: common.ARGO_ROLLOUT, NamespaceObjects: unhealthyRol, RestartConfig: cfg(0)},
	}
}

func TestProcessRestartsStaggered_WithinNamespace_DeploymentsThenRollouts_GatedThenUngated(t *testing.T) {
	rec := installRecorders(t, nil, 0)
	k8sFake := aliasK8sFake{aliases: map[string]string{"ns": ""}}

	ReconcilerWrapper{}.ProcessRestartsStaggered(k8sFake, k8s.ClientSet{}, buildRestarts(
		map[string][]string{"ns": {"dep-h1", "dep-h2"}},
		map[string][]string{"ns": {"rol-h1"}},
		map[string][]string{"ns": {"dep-u1"}},
		map[string][]string{"ns": {"rol-u1"}},
	))

	want := []string{
		"restart-dep:ns/dep-h1", "wait-dep:ns/dep-h1",
		"restart-dep:ns/dep-h2", "wait-dep:ns/dep-h2",
		"restart-dep:ns/dep-u1", // unhealthy: no wait (master behavior)
		"restart-rol:ns/rol-h1", "wait-rol:ns/rol-h1",
		"restart-rol:ns/rol-u1",
	}
	if !reflect.DeepEqual(rec.sequence, want) {
		t.Errorf("unexpected sequence\nwant %v\ngot  %v", want, rec.sequence)
	}
}

func TestProcessRestartsStaggered_SameAliasSequential_DifferentAliasParallel(t *testing.T) {
	aliases := map[string]string{
		"qbo-a": "Intuit.qbo", "qbo-b": "Intuit.qbo", "qbo-c": "Intuit.qbo",
		"pay-a": "Intuit.payments", "pay-b": "Intuit.payments",
		"solo": "", // no annotation: its own group
	}
	nsToGroup := map[string]string{}
	for ns, a := range aliases {
		if a != "" {
			nsToGroup[ns] = "alias:" + a
		}
	}
	rec := installRecorders(t, nsToGroup, 20*time.Millisecond)
	common.StaggerMaxParallelGroups = 5

	healthy := map[string][]string{}
	for ns := range aliases {
		healthy[ns] = []string{"d1", "d2"}
	}
	ReconcilerWrapper{}.ProcessRestartsStaggered(aliasK8sFake{aliases: aliases}, k8s.ClientSet{}, buildRestarts(healthy, nil, nil, nil))

	if len(rec.groupViolations) > 0 {
		t.Errorf("namespaces of the same alias overlapped: %v", rec.groupViolations)
	}
	if rec.maxActive < 2 {
		t.Errorf("expected different alias groups to run in parallel, max concurrency was %d", rec.maxActive)
	}
	if rec.maxActive > 3 {
		t.Errorf("only 3 groups exist, but max concurrency was %d", rec.maxActive)
	}

	// Within the qbo group, namespaces must be visited in sorted order, fully one after another.
	var qboOrder []string
	for _, call := range rec.sequence {
		if strings.HasPrefix(call, "restart-dep:qbo-") {
			ns := strings.TrimPrefix(strings.SplitN(call, "/", 2)[0], "restart-dep:")
			if len(qboOrder) == 0 || qboOrder[len(qboOrder)-1] != ns {
				qboOrder = append(qboOrder, ns)
			}
		}
	}
	if want := []string{"qbo-a", "qbo-b", "qbo-c"}; !reflect.DeepEqual(qboOrder, want) {
		t.Errorf("qbo group order: want %v, got %v", want, qboOrder)
	}
	if got := len(rec.sequence); got != 12*2 { // 6 ns * 2 deployments * (restart + wait)
		t.Errorf("expected 24 recorded calls, got %d", got)
	}
}

func TestProcessRestartsStaggered_RespectsParallelCap(t *testing.T) {
	aliases := map[string]string{"a": "", "b": "", "c": "", "d": "", "e": "", "f": ""} // 6 groups
	rec := installRecorders(t, nil, 20*time.Millisecond)
	common.StaggerMaxParallelGroups = 2

	healthy := map[string][]string{}
	for ns := range aliases {
		healthy[ns] = []string{"d1"}
	}
	ReconcilerWrapper{}.ProcessRestartsStaggered(aliasK8sFake{aliases: aliases}, k8s.ClientSet{}, buildRestarts(healthy, nil, nil, nil))

	if rec.maxActive > 2 {
		t.Errorf("cap of 2 exceeded: max concurrency %d", rec.maxActive)
	}
	if rec.maxActive < 2 {
		t.Errorf("expected cap to allow 2 concurrent groups, got %d", rec.maxActive)
	}
	if len(rec.sequence) != 12 {
		t.Errorf("expected all 6 namespaces processed (12 calls), got %d", len(rec.sequence))
	}
}

func TestProcessRestartsStaggered_NamespaceListFailure_EachNamespaceOwnGroup(t *testing.T) {
	rec := installRecorders(t, nil, 10*time.Millisecond)
	common.StaggerMaxParallelGroups = 5

	healthy := map[string][]string{"x": {"d1"}, "y": {"d1"}, "z": {"d1"}}
	ReconcilerWrapper{}.ProcessRestartsStaggered(aliasK8sFake{listErr: errors.New("forbidden")}, k8s.ClientSet{}, buildRestarts(healthy, nil, nil, nil))

	if len(rec.sequence) != 6 {
		t.Errorf("expected all 3 namespaces processed despite alias lookup failure, got %d calls", len(rec.sequence))
	}
	if rec.maxActive < 2 {
		t.Errorf("expected namespaces to run as independent groups in parallel, max concurrency %d", rec.maxActive)
	}
}

func TestProcessRestartsStaggered_NothingToDo(t *testing.T) {
	rec := installRecorders(t, nil, 0)
	ReconcilerWrapper{}.ProcessRestartsStaggered(aliasK8sFake{}, k8s.ClientSet{}, buildRestarts(nil, nil, nil, nil))
	if len(rec.sequence) != 0 {
		t.Errorf("expected no calls, got %v", rec.sequence)
	}
}

func TestGroupByAlias(t *testing.T) {
	byNS := map[string]*staggerNamespace{
		"qbo-b": {name: "qbo-b"}, "qbo-a": {name: "qbo-a"}, "solo": {name: "solo"}, "pay": {name: "pay"},
	}
	aliases := map[string]string{"qbo-a": "Intuit.qbo", "qbo-b": "Intuit.qbo", "pay": "Intuit.payments"}

	groups := groupByAlias(byNS, aliases)

	var keys []string
	got := map[string][]string{}
	for _, g := range groups {
		keys = append(keys, g.key)
		for _, ns := range g.namespaces {
			got[g.key] = append(got[g.key], ns.name)
		}
	}
	wantKeys := []string{"alias:Intuit.payments", "alias:Intuit.qbo", "namespace:solo"}
	if !reflect.DeepEqual(keys, wantKeys) {
		t.Errorf("group keys: want %v, got %v", wantKeys, keys)
	}
	if !reflect.DeepEqual(got["alias:Intuit.qbo"], []string{"qbo-a", "qbo-b"}) {
		t.Errorf("qbo group namespaces not sorted: %v", got["alias:Intuit.qbo"])
	}
	if !sort.StringsAreSorted(keys) {
		t.Errorf("groups not sorted: %v", keys)
	}
}

func TestNamespaceAliases_IgnoresBlankAnnotation(t *testing.T) {
	fake := aliasK8sFake{aliases: map[string]string{"a": "Intuit.a", "b": "   ", "c": ""}}
	got := namespaceAliases(fake, k8s.ClientSet{})
	if want := map[string]string{"a": "Intuit.a"}; !reflect.DeepEqual(got, want) {
		t.Errorf("want %v, got %v", want, got)
	}
}

// ProcessRestart must dispatch to the staggered path when the flag is on and
// leave the per-type Restart() loop alone when it is off.
func TestProcessRestart_DispatchesOnFlag(t *testing.T) {
	t.Run("flag on uses staggered path", func(t *testing.T) {
		rec := installRecorders(t, nil, 0)
		config := BuildTestFlippyConfig()
		config.Spec.PostFilterRestarts = nil
		config.Spec.ImageList = []string{"foo", "boo"} // force drift so the fake reports restarts
		err := ReconcilerWrapper{}.ProcessRestart(aliasK8sFake{aliases: map[string]string{HappyPreCondition: ""}}, BuildTestK8SClientSet(), config)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, call := range rec.sequence {
			if strings.HasPrefix(call, "UNEXPECTED") {
				t.Fatalf("flag on must not use Restart(): %v", rec.sequence)
			}
		}
		if len(rec.sequence) == 0 {
			t.Fatalf("expected staggered restarts to be recorded")
		}
	})

	t.Run("flag off uses Restart()", func(t *testing.T) {
		rec := installRecorders(t, nil, 0)
		common.StaggerNamespacesEnabled = false
		config := BuildTestFlippyConfig()
		config.Spec.PostFilterRestarts = nil
		config.Spec.ImageList = []string{"foo", "boo"}
		if err := (ReconcilerWrapper{}).ProcessRestart(fakeK8sApi, BuildTestK8SClientSet(), config); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rec.sequence) == 0 {
			t.Fatalf("expected Restart() calls to be recorded")
		}
		for _, call := range rec.sequence {
			if !strings.HasPrefix(call, "UNEXPECTED Restart()") {
				t.Fatalf("flag off must only use Restart(): %v", rec.sequence)
			}
		}
	})
}
