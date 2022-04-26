package k8s

import (
	"k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestCreateNetworkPolicy(t *testing.T) {

	k8 := K8sWrapper{}

	networkPolicyToCreate := &v1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "TestNetworkPolicy",
			Namespace: "TestNamespace",
		},
	}

	err := k8.CreateNetworkPolicy(fake.NewSimpleClientset(), networkPolicyToCreate, "TestNamespace")
	if err != nil {
		t.Errorf("CreateNetworkPolicy test failed, expected no error. Actual error = %v", err)
		return
	}
}
