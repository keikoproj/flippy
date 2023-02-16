package k8s

import (
	"context"
	"github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	argo "github.com/argoproj/argo-rollouts/pkg/client/clientset/versioned"
	"github.com/keikoproj/flippy/pkg/k8s-utils/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (K8sWrapper) RolloutRestartArgoRollouts(kubeconfigpath string, namespace string, argoRolloutName string) (string, error) {

	var args []string
	args = append(args, "argo", "rollouts", "--kubeconfig="+kubeconfigpath, "-n", namespace, "restart", argoRolloutName)
	return K8s.ExecuteKubectlCommand(args)
}

func (K8sWrapper) RolloutArogRolloutStatus(kubeconfigpath string, namespace string, argoRolloutName string) (string, error) {

	var args []string
	args = append(args, "argo", "rollouts", "--kubeconfig="+kubeconfigpath, "status", "-n", namespace, argoRolloutName, "--watch=false")
	return K8s.ExecuteKubectlCommand(args)
}

func (K8sWrapper) GetArgoRollouts(clientset argo.Interface, namespace string) (*v1alpha1.RolloutList, error) {

	rolloutClient := clientset.ArgoprojV1alpha1().Rollouts(namespace)
	rolloutList, err := rolloutClient.List(context.TODO(), metav1.ListOptions{})
	return rolloutList, err
}

func (K8sWrapper) GetArgoRolloutsWithSpecAnnotationFilter(clientset argo.Interface, namespace string, annotations map[string]string) ([]string, error) {
	argoRolloutList := make([]string, 0)
	argoRollouts, err := K8s.GetArgoRollouts(clientset, namespace)

	if err != nil {
		return argoRolloutList, err
	}

	for _, argoRollout := range argoRollouts.Items {
		if utils.IsStringMapSubset(argoRollout.Spec.Template.ObjectMeta.Annotations, annotations) {
			argoRolloutList = append(argoRolloutList, argoRollout.Name)
		}
	}

	return argoRolloutList, nil
}
