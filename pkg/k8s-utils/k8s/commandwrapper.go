package k8s

import (
	"os/exec"
	"strconv"
	"strings"
)

func (K8sWrapper) PatchResource(kubeconfigpath string, namespace string, resource string, resourceName string, patchJson string) error {
	_, err := exec.Command("kubectl", "patch", "--kubeconfig="+kubeconfigpath, "-n", namespace, resource, resourceName, "-p", patchJson, "--type", "json").CombinedOutput()

	return err
}

func (K8sWrapper) ScaleDeployment(kubeconfigpath string, namespace string, deploymentname string, scale int) error {
	_, err := exec.Command("kubectl", "scale", "--kubeconfig="+kubeconfigpath, "-n", namespace, "deployment/"+deploymentname,
		"--replicas="+strconv.Itoa(scale)).CombinedOutput()

	return err
}

func (K8sWrapper) ApplyYaml(kubeconfigpath string, yamlFilePath string) error {
	_, err := exec.Command("kubectl", "--kubeconfig="+kubeconfigpath, "apply", "-f", yamlFilePath).CombinedOutput()
	return err
}

func (K8sWrapper) DeleteYaml(kubeconfigpath string, yamlFilePath string) error {
	_, err := exec.Command("kubectl", "--kubeconfig="+kubeconfigpath, "delete", "-f", yamlFilePath).CombinedOutput()
	return err
}

func (K8sWrapper) GetServiceEntries(kubeconfigpath string, namespace string) (map[string]string, error) {
	var serviceEntriesMap map[string]string
	serviceEntriesMap = make(map[string]string)
	output, err := exec.Command("kubectl", "--kubeconfig="+kubeconfigpath, "get", "se", "-n", namespace).CombinedOutput()
	if err != nil {
		return serviceEntriesMap, err
	}
	splits := strings.SplitAfter(string(output), "\n")
	for i, split := range splits {
		if i > 0 && len(split) > 0 {
			split_part1 := strings.SplitAfter(split, "[")
			split_part2 := strings.SplitAfter(split_part1[1], "]")
			serviceEntryName := split_part1[0]
			serviceEntryHostName := split_part2[0]

			serviceEntryName = strings.TrimSpace(strings.ReplaceAll(serviceEntryName, "[", ""))
			serviceEntryHostName = strings.TrimSpace(strings.ReplaceAll(serviceEntryHostName, "]", ""))

			serviceEntriesMap[serviceEntryName] = serviceEntryHostName
		}
	}
	return serviceEntriesMap, nil
}

func (K8sWrapper) RestartContainer(kubeconfigpath string, namespace string, podname string, containerName string) (string, error) {
	//This is achieved by killing container process. K8S will start container due to deployment spec.

	output, err := exec.Command("kubectl", "exec", "-it", podname, "--kubeconfig="+kubeconfigpath, "-n", namespace, "-c", containerName, "--", "/bin/bash", "-c", "kill 1").CombinedOutput()

	if err != nil || strings.Contains(string(output), "Unable to use a TTY") {
		return "", err
	} else {
		return string(output), err
	}
}

func (K8sWrapper) RolloutRestartDeployment(kubeconfigpath string, namespace string, deploymentName string) (string, error) {
	output, err := exec.Command("kubectl", "--kubeconfig="+kubeconfigpath, "-n", namespace, "rollout", "restart", "deployment", deploymentName).CombinedOutput()

	if err != nil || strings.Contains(string(output), "Unable to use a TTY") {
		return "", err
	} else {
		return string(output), nil
	}
}

func (K8sWrapper) RolloutDeploymentStatus(kubeconfigpath string, namespace string, deploymentName string) (string, error) {

	var args []string
	args = append(args, "--kubeconfig="+kubeconfigpath, "-n", namespace, "rollout", "status", "deployment", deploymentName, "--watch=false")
	output, err := exec.Command("kubectl", args...).CombinedOutput()

	if err != nil || strings.Contains(string(output), "Unable to use a TTY") {
		return string(output), err
	} else {
		return string(output), nil
	}
}
