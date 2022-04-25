package k8s

import (
	"context"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func (K8sWrapper) CreateJob(clientset kubernetes.Interface, job *batchv1.Job, namespace string) error {
	jobsclient := clientset.BatchV1().Jobs(namespace)
	_, err := jobsclient.Create(context.TODO(), job, metav1.CreateOptions{})
	if err != nil {
		_, err := jobsclient.Create(context.TODO(), job, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (K8sWrapper) DeleteJob(clientset kubernetes.Interface, namespace string, jobName string) error {
	jobsclient := clientset.BatchV1().Jobs(namespace)
	err := jobsclient.Delete(context.TODO(), jobName, metav1.DeleteOptions{})
	if err != nil {
		err := jobsclient.Delete(context.TODO(), jobName, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (K8sWrapper) DeleteJobWithPods(clientset kubernetes.Interface, namespace string, jobName string) error {
	err := K8s.DeleteJob(clientset, namespace, jobName)
	if err != nil {
		return err
	}

	err = K8s.DeletePods(clientset, namespace, jobName)
	if err != nil {
		return err
	}

	return nil
}
