package k8s

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func (K8sWrapper) PatchResource(kubeconfigpath string, namespace string, resource string, resourceName string, patchJson string) (string, error) {
	var args []string
	args = append(args, "patch", "--kubeconfig="+kubeconfigpath, "-n", namespace, resource, resourceName, "-p", patchJson, "--type", "json")
	return K8s.ExecuteKubectlCommand(args)
}

func (K8sWrapper) ScaleDeployment(kubeconfigpath string, namespace string, deploymentname string, scale int) (string, error) {

	var args []string
	args = append(args, "scale", "--kubeconfig="+kubeconfigpath, "-n", namespace, "deployment/"+deploymentname,
		"--replicas="+strconv.Itoa(scale))

	return K8s.ExecuteKubectlCommand(args)
}

func (K8sWrapper) ApplyYaml(kubeconfigpath string, yamlFilePath string) (string, error) {
	var args []string
	args = append(args, "--kubeconfig="+kubeconfigpath, "apply", "-f", yamlFilePath)
	return K8s.ExecuteKubectlCommand(args)
}

func (K8sWrapper) DeleteYaml(kubeconfigpath string, yamlFilePath string) (string, error) {
	var args []string
	args = append(args, "--kubeconfig="+kubeconfigpath, "delete", "-f", yamlFilePath)
	return K8s.ExecuteKubectlCommand(args)
}

func (K8sWrapper) GetServiceEntries(kubeconfigpath string, namespace string) (map[string]string, error) {
	var serviceEntriesMap map[string]string
	serviceEntriesMap = make(map[string]string)
	var args []string
	args = append(args, "--kubeconfig="+kubeconfigpath, "get", "se", "-n", namespace)
	output, err := K8s.ExecuteKubectlCommand(args)
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

	var args []string
	args = append(args, "exec", "-it", podname, "--kubeconfig="+kubeconfigpath, "-n", namespace, "-c", containerName, "--", "/bin/bash", "-c", "kill 1")

	return K8s.ExecuteKubectlCommand(args)
}

func (K8sWrapper) RolloutRestartDeployment(kubeconfigpath string, namespace string, deploymentName string) (string, error) {
	var args []string
	args = append(args, "--kubeconfig="+kubeconfigpath, "-n", namespace, "rollout", "restart", "deployment", deploymentName)

	return K8s.ExecuteKubectlCommand(args)
}

func (K8sWrapper) RolloutDeploymentStatus(kubeconfigpath string, namespace string, deploymentName string) (string, error) {

	var args []string
	args = append(args, "--kubeconfig="+kubeconfigpath, "-n", namespace, "rollout", "status", "deployment", deploymentName, "--watch=false")

	return K8s.ExecuteKubectlCommand(args)
}

func (K8sWrapper) ExecuteKubectlCommand(cmdParameter []string) (string, error) {

	output, err := exec.Command("kubectl", cmdParameter...).CombinedOutput()

	if err != nil {
		err = errors.New(fmt.Sprintf("Failed to execute - kubectl %s, Console Output - %s", strings.Join(cmdParameter, " "), output))
	}

	return string(output), err

}
