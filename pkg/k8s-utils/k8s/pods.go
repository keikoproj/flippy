package k8s

import (
	"bytes"
	"context"
	log "github.com/sirupsen/logrus"
	"io"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
	. "time"
)

func (K8sWrapper) GetRunningPods(clientset kubernetes.Interface, namespace string, podNameContains string) ([]corev1.Pod, error) {
	var returnPodList []corev1.Pod

	podList, err := K8s.GetAllPodsInNamespace(clientset, namespace)
	for err != nil {
		log.Info("Retrying get all pod in namespaces. Error - ", err)
		Sleep(1 * Minute)
		podList, err = K8s.GetAllPodsInNamespace(clientset, namespace)
	}

	for _, pod := range podList {
		if strings.Contains(pod.ObjectMeta.Name, podNameContains) && pod.Status.Phase == "Running" {
			returnPodList = append(returnPodList, pod)
		}
	}

	return returnPodList, err
}

func (K8sWrapper) GetAllPodsInNamespace(clientset kubernetes.Interface, namespace string) ([]corev1.Pod, error) {
	var returnPodList []corev1.Pod

	podList, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		return returnPodList, err
	}

	for _, pod := range podList.Items {
		returnPodList = append(returnPodList, pod)
	}

	return returnPodList, err
}

func (K8sWrapper) DeletePod(clientset kubernetes.Interface, namespace string, podName string) error {

	podclient := clientset.CoreV1().Pods(namespace)
	err := podclient.Delete(context.TODO(), podName, metav1.DeleteOptions{})
	if err != nil {
		err := podclient.Delete(context.TODO(), podName, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}
func (K8sWrapper) DeletePodsWithRetry(clientset kubernetes.Interface, namespace string, podNameContains string) error {
	podList, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		return err
	}

	for _, pod := range podList.Items {
		if strings.Contains(pod.ObjectMeta.Name, podNameContains) {
			err = K8s.DeletePod(clientset, namespace, pod.Name)
			for err != nil {
				log.Info("Retrying pod deletion. Namespace -"+namespace+", PodName - "+podNameContains, err)
				err = K8s.DeletePod(clientset, namespace, pod.Name)
			}
		}
	}

	return nil
}
func (K8sWrapper) DeletePods(clientset kubernetes.Interface, namespace string, podNameContains string) error {
	err := K8s.DeletePodsWithRetry(clientset, namespace, podNameContains)
	for i := 0; err != nil && i < 5; i++ {
		err = K8s.DeletePodsWithRetry(clientset, namespace, podNameContains)
	}
	return err
}

func (K8sWrapper) GetPodLogs(clientset kubernetes.Interface, podName string, namespace string, container string, fromTime Time) (string, error) {
	logsAfter := metav1.NewTime(fromTime)

	logOptions := corev1.PodLogOptions{
		Container: container,
		Follow:    false,
		SinceTime: &logsAfter,
	}

	logReq := clientset.CoreV1().Pods(namespace).GetLogs(podName, &logOptions)

	podLog, err := logReq.Stream(context.TODO())
	if err != nil {
		return "", err
	}

	defer podLog.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLog)
	if err != nil {
		return "", err
	}

	str := buf.String()
	return str, err
}

func (K8sWrapper) RestartContainers(clientset kubernetes.Interface, kubeconfigpath string, namespace string, podnamecontains string, containerName string) (string, error) {

	podList, err := K8s.GetRunningPods(clientset, namespace, podnamecontains)
	if err != nil {
		return "", err
	}

	for _, pod := range podList {
		return K8s.RestartContainer(kubeconfigpath, namespace, pod.ObjectMeta.Name, containerName)
	}
	return "", nil
}
func (K8sWrapper) GetLogFromFirstPod(clientset kubernetes.Interface, namespace string, podNameContains string, containerName string, logsFromTime Time) (string, error) {

	var logStr string
	//Assert log is going
	podList, err := K8s.GetRunningPods(clientset, namespace, podNameContains)
	if err != nil {
		return logStr, err
	}

	if len(podList) > 0 {
		podName := podList[0].ObjectMeta.Name
		log.Info("Fetching log for pod - ", podName)
		logStr, err = K8s.GetPodLogs(clientset, podName, namespace, containerName, logsFromTime)
		if err != nil {
			return logStr, err
		}
	}
	return logStr, nil
}
