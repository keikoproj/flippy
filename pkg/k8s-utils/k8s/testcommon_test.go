package k8s

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func BuildTestK8SClientSet() ClientSet {
	k8sClientSet := fake.NewSimpleClientset(&corev1.Pod{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		//Spec:       ,
		//Status:     ,
	})

	testClientSet := ClientSet{
		K8sClientSet:         k8sClientSet,
		ArgoRolloutClientSet: nil,
	}

	return testClientSet
}
