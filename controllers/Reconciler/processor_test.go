package Reconciler

import (
	v1 "keikoproj.intuit.com/Flippy/api/v1"
	"keikoproj.intuit.com/Flippy/pkg/common"
	"keikoproj.intuit.com/Flippy/pkg/k8s-utils/k8s"
	"testing"
)

func TestReconcilerWrapper_ProcessNamespaceRestarts(t *testing.T) {
	type testargs struct {
		k8s            k8s.K8sAPI
		restartObjects []common.RestartObjects
	}

	namespaceTestMapHappy := make(map[string][]string)
	namespaceTestMapHappy[HappyPreCondition] = append(namespaceTestMapHappy[HappyPreCondition], HappyPreCondition)

	namespaceTestMapNotHappy := make(map[string][]string)
	namespaceTestMapHappy[NotHappyPreCondition] = append(namespaceTestMapHappy[NotHappyPreCondition], NotHappyPreCondition)

	tests := []struct {
		name string
		args testargs
	}{
		{"Deployment", testargs{fakeK8sApi, []common.RestartObjects{
			common.RestartObjects{
				Type:             common.DEPLOYMENT,
				NamespaceObjects: namespaceTestMapHappy,
				RestartConfig: v1.StatusCheckConfig{
					CheckStatus:   true,
					MaxRetry:      0,
					RetryDuration: 0,
				},
			}}}},
		{"ArgoRollout", testargs{fakeK8sApi, []common.RestartObjects{
			common.RestartObjects{
				Type:             common.ARGO_ROLLOUT,
				NamespaceObjects: namespaceTestMapHappy,
				RestartConfig: v1.StatusCheckConfig{
					CheckStatus:   true,
					MaxRetry:      0,
					RetryDuration: 0,
				},
			}}}},
		{"DeploymentAndArgo", testargs{fakeK8sApi, []common.RestartObjects{
			common.RestartObjects{
				Type:             common.DEPLOYMENT,
				NamespaceObjects: namespaceTestMapNotHappy,
				RestartConfig: v1.StatusCheckConfig{
					CheckStatus:   true,
					MaxRetry:      0,
					RetryDuration: 0,
				},
			}, common.RestartObjects{
				Type:             common.ARGO_ROLLOUT,
				NamespaceObjects: namespaceTestMapNotHappy,
				RestartConfig: v1.StatusCheckConfig{
					CheckStatus:   false,
					MaxRetry:      0,
					RetryDuration: 0,
				},
			}}}},
	}

	test := ReconcilerWrapper{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.ProcessNamespaceRestarts(tt.args.k8s, tt.args.restartObjects)
		})
	}
}

func TestReconcilerWrapper_ProcessRestart(t *testing.T) {
	type args struct {
		k8s       k8s.K8sAPI
		clientset k8s.ClientSet
		config    v1.FlippyConfig
	}

	var nonhappyPostFilterRestartsTests []v1.FlippyCondition
	nonhappyPostFilterRestartsTests = append(nonhappyPostFilterRestartsTests, v1.FlippyCondition{
		K8S: v1.K8S{
			Type:      common.DEPLOYMENT,
			Name:      NotHappyPreCondition,
			Namespace: NotHappyPreCondition,
		},
		Status:            "",
		StatusCheckConfig: BuildTestFlippyStatusCheckConfig(),
	})

	nonhappyPostFilterRestartsTests = append(nonhappyPostFilterRestartsTests, v1.FlippyCondition{
		K8S: v1.K8S{
			Type:      common.ARGO_ROLLOUT,
			Name:      NotHappyPreCondition,
			Namespace: NotHappyPreCondition,
		},
		Status:            "",
		StatusCheckConfig: BuildTestFlippyStatusCheckConfig(),
	})

	flippyConfigTest := BuildTestFlippyConfig()

	testClientSet := BuildTestK8SClientSet()

	testArg1 := args{
		k8s:       fakeK8sApi,
		clientset: testClientSet,
		config:    flippyConfigTest,
	}

	testArg2 := args{
		k8s:       fakeK8sApi,
		clientset: testClientSet,
		config:    flippyConfigTest,
	}

	testArg2.config.Spec.PostFilterRestarts = nonhappyPostFilterRestartsTests

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"HappyPathDeploymentAndArgoRollout", testArg1, false},
		{"NonHappyPathDeploymentAndArgoRollout", testArg2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := ReconcilerWrapper{}
			err := re.ProcessRestart(tt.args.k8s, tt.args.clientset, tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessRestart() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
