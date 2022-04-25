package Reconciler

import (
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	crdv1 "keikoproj.intuit.com/Flippy/api/v1"
	"keikoproj.intuit.com/Flippy/pkg/common"
	"keikoproj.intuit.com/Flippy/pkg/k8s-utils/k8s"
	"testing"
)

func TestReconcilerWrapper_IsProxyVersionChange(t *testing.T) {
	type testargs struct {
		imageList []string
	}
	tests := []struct {
		name string
		args testargs
		want bool
	}{
		{"Empty", testargs{imageList: []string{}}, false},
		{"FirstCall", testargs{imageList: []string{"foo", "bar"}}, true},
		{"DockerImageChange1", testargs{imageList: []string{"zoo", "car"}}, true},
		{"DockerImageChange2", testargs{imageList: []string{"zoo", "foo", "car"}}, true},
		{"NoChange", testargs{imageList: []string{"zoo", "foo", "car"}}, false},
	}
	test := ReconcilerWrapper{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := test.IsProxyVersionChange(tt.args.imageList); got != tt.want {
				t.Errorf("IsProxyVersionChange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReconcilerWrapper_IsPreConditionSatisfied(t *testing.T) {

	mockDeployment := appsv1.Deployment{}

	deploymentPrecondition := crdv1.FlippyCondition{
		K8S: crdv1.K8S{
			Type:      common.DEPLOYMENT,
			Name:      HappyPreCondition,
			Namespace: HappyPreCondition,
		},
		Status: "",
		StatusCheckConfig: crdv1.StatusCheckConfig{
			CheckStatus:   true,
			MaxRetry:      1,
			RetryDuration: 10,
		},
	}

	deploymentNotHappyPrecondition := crdv1.FlippyCondition{
		K8S: crdv1.K8S{
			Type:      common.DEPLOYMENT,
			Name:      NotHappyPreCondition,
			Namespace: NotHappyPreCondition,
		},
		Status: "",
		StatusCheckConfig: crdv1.StatusCheckConfig{
			CheckStatus:   true,
			MaxRetry:      1,
			RetryDuration: 10,
		},
	}

	argoPrecondition := crdv1.FlippyCondition{
		K8S: crdv1.K8S{
			Type:      common.ARGO_ROLLOUT,
			Name:      HappyPreCondition,
			Namespace: HappyPreCondition,
		},
		Status: "",
		StatusCheckConfig: crdv1.StatusCheckConfig{
			CheckStatus:   true,
			MaxRetry:      1,
			RetryDuration: 10,
		},
	}

	mockDeployment.Name = deploymentPrecondition.K8S.Name
	mockDeployment.Namespace = deploymentPrecondition.K8S.Namespace

	type testParam struct {
		k8sapi     k8s.K8sAPI
		clientset  kubernetes.Interface
		condition  crdv1.FlippyCondition
		retryCount int
	}
	tests := []struct {
		name string
		args testParam
		want bool
	}{
		{HappyPreCondition + common.DEPLOYMENT, testParam{
			k8sapi:     fakeK8sApi,
			clientset:  fake.NewSimpleClientset(),
			condition:  deploymentPrecondition,
			retryCount: 0,
		}, true},
		{"NonImplemented" + common.ARGO_ROLLOUT, testParam{
			k8sapi:     fakeK8sApi,
			clientset:  fake.NewSimpleClientset(),
			condition:  argoPrecondition,
			retryCount: 0,
		}, true},
		{NotHappyPreCondition + common.DEPLOYMENT, testParam{
			k8sapi:     fakeK8sApi,
			clientset:  fake.NewSimpleClientset(),
			condition:  deploymentNotHappyPrecondition,
			retryCount: 0,
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := ReconcilerWrapper{}
			if got := re.IsPreConditionSatisfied(tt.args.k8sapi, tt.args.clientset, tt.args.condition, tt.args.retryCount); got != tt.want {
				t.Errorf("IsPreConditionSatisfied() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReconcilerWrapper_Handle(t *testing.T) {
	type args struct {
		req              crdv1.FlippyConfig
		clientsetWrapper k8s.ClientSet
		k8sUtils         k8s.K8sAPI
	}

	negativeFlippyConfig := BuildTestFlippyConfig()
	negativeFlippyConfig.Spec.ProcessFilter.NamespaceLabels[NotHappyPreCondition] = NotHappyPreCondition

	preConditionFlippy := BuildTestFlippyConfig()
	testPreConditionFlippyConfig := BuildTestFlippyStatusCheckConfig()
	testPreConditionFlippyConfig.CheckStatus = true

	preConditionFlippy.Spec.Preconditions[0].StatusCheckConfig = testPreConditionFlippyConfig

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"HappyTest", args{
			req:              BuildTestFlippyConfig(),
			clientsetWrapper: BuildTestK8SClientSet(),
			k8sUtils:         fakeK8sApi,
		}, false},
		{"NegativeTest", args{
			req:              negativeFlippyConfig,
			clientsetWrapper: BuildTestK8SClientSet(),
			k8sUtils:         fakeK8sApi,
		}, true},
		{"StatusCheck", args{
			req:              preConditionFlippy,
			clientsetWrapper: BuildTestK8SClientSet(),
			k8sUtils:         fakeK8sApi,
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := ReconcilerWrapper{}
			//Reset Docker Image cache
			dockerImages = []string{}
			err := re.Handle(tt.args.req, tt.args.clientsetWrapper, tt.args.k8sUtils)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
