package utils

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/argoproj/argo-rollouts/pkg/client/clientset/versioned"
	"github.com/keikoproj/flippy/pkg/common"
	"github.com/keikoproj/flippy/pkg/k8s-utils/config"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

func ConvertYamlToK8s(yamlFile []byte) (runtime.Object, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode

	//Decode
	obj, _, err := decode([]byte(yamlFile), nil, nil)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func ReadK8sFromFile(path string) (runtime.Object, error) {
	filepath, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	filepath = filepath + "/" + path
	log.Println("Reading from ", filepath)
	//Read Yaml File
	yamlFile, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	obj, err := ConvertYamlToK8s(yamlFile)
	if err != nil {
		return nil, err
	}
	return obj, err
}

func Setup() (*rest.Config, *kubernetes.Clientset, error) {

	if common.KubeconfigPath == "" {
		return config.SetupInCluster()
	} else {
		return config.Setup(common.KubeconfigPath)
	}
}

func SetupArgoRollout() (*rest.Config, *versioned.Clientset, error) {
	if common.KubeconfigPath == "" {
		return config.SetupInClusterArgoRollout()
	} else {
		return config.SetupArgoRollout(common.KubeconfigPath)
	}
}

func StringArrayContains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func IsStringMapSubset(masterMap map[string]string, subsetMap map[string]string) bool {
	flippyIgnore, ok := masterMap[common.IgnoreMetadata]

	if ok && strings.ToLower(flippyIgnore) == "true" {
		return false
	}

	match := 0
	for key, value := range subsetMap {
		masterValue, ok := masterMap[key]
		if ok && masterValue == value {
			match++
		}
	}

	if match == len(subsetMap) {
		return true
	}
	return false
}
