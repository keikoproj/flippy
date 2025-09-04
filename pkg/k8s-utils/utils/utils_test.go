package utils

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/keikoproj/flippy/pkg/common"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestIsStringMapSubset(t *testing.T) {

	masterMap := make(map[string]string)
	masterMap["test1"] = "test"
	masterMap["test2"] = ""
	masterMap["test3"] = "true"
	masterMap["test4"] = "false"

	subsetMap := make(map[string]string)
	subsetMap["test1"] = "test"

	tests := []struct {
		name string
		args string
		want bool
	}{
		{"No addition to label", "empty", true},
		{"Addition to label to ignore flippy with empty value", "test2", true},
		{"Addition to label to ignore flippy", "test3", false},
		{"Addition to label to ignore flippy with default value", "test4", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args != "empty" {
				common.IgnoreMetadataKey = tt.args
			}
			if got := IsStringMapSubset(masterMap, subsetMap); got != tt.want {
				t.Errorf("IsStringMapSubset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsStringMapSubsetNegative(t *testing.T) {

	masterMap := make(map[string]string)
	masterMap["test1"] = "test"
	masterMap["test2"] = ""
	masterMap["test3"] = "true"
	masterMap["test4"] = "false"

	subsetMap := make(map[string]string)
	subsetMap["test5"] = "test"

	got := IsStringMapSubset(masterMap, subsetMap)
	assert.Equal(t, false, got)
}

func TestIsStringMapSubsetAdditional(t *testing.T) {
	tests := []struct {
		name              string
		masterMap         map[string]string
		subsetMap         map[string]string
		ignoreMetadataKey string
		expected          bool
		description       string
	}{
		{
			name:              "Empty maps",
			masterMap:         map[string]string{},
			subsetMap:         map[string]string{},
			ignoreMetadataKey: "",
			expected:          true,
			description:       "Empty subset should match empty master",
		},
		{
			name:              "Empty subset with non-empty master",
			masterMap:         map[string]string{"key1": "value1"},
			subsetMap:         map[string]string{},
			ignoreMetadataKey: "",
			expected:          true,
			description:       "Empty subset should match any master",
		},
		{
			name:              "Partial match",
			masterMap:         map[string]string{"key1": "value1", "key2": "value2", "key3": "value3"},
			subsetMap:         map[string]string{"key1": "value1", "key2": "value2"},
			ignoreMetadataKey: "",
			expected:          true,
			description:       "Subset with all matching keys should return true",
		},
		{
			name:              "Partial match with one mismatch",
			masterMap:         map[string]string{"key1": "value1", "key2": "value2", "key3": "value3"},
			subsetMap:         map[string]string{"key1": "value1", "key2": "different"},
			ignoreMetadataKey: "",
			expected:          false,
			description:       "Subset with one mismatched value should return false",
		},
		{
			name:              "Case sensitive ignore key - uppercase TRUE",
			masterMap:         map[string]string{"ignore-key": "TRUE", "key1": "value1"},
			subsetMap:         map[string]string{"key1": "value1"},
			ignoreMetadataKey: "ignore-key",
			expected:          false,
			description:       "Uppercase TRUE should be converted to lowercase and trigger ignore",
		},
		{
			name:              "Case sensitive ignore key - mixed case True",
			masterMap:         map[string]string{"ignore-key": "True", "key1": "value1"},
			subsetMap:         map[string]string{"key1": "value1"},
			ignoreMetadataKey: "ignore-key",
			expected:          false,
			description:       "Mixed case True should be converted to lowercase and trigger ignore",
		},
		{
			name:              "Ignore key with false value",
			masterMap:         map[string]string{"ignore-key": "false", "key1": "value1"},
			subsetMap:         map[string]string{"key1": "value1"},
			ignoreMetadataKey: "ignore-key",
			expected:          true,
			description:       "Ignore key with false value should not trigger ignore",
		},
		{
			name:              "Ignore key with empty value",
			masterMap:         map[string]string{"ignore-key": "", "key1": "value1"},
			subsetMap:         map[string]string{"key1": "value1"},
			ignoreMetadataKey: "ignore-key",
			expected:          true,
			description:       "Ignore key with empty value should not trigger ignore",
		},
		{
			name:              "Large maps with all matches",
			masterMap:         map[string]string{"k1": "v1", "k2": "v2", "k3": "v3", "k4": "v4", "k5": "v5"},
			subsetMap:         map[string]string{"k1": "v1", "k3": "v3", "k5": "v5"},
			ignoreMetadataKey: "",
			expected:          true,
			description:       "Large maps with all subset keys matching should return true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the global ignore metadata key
			originalIgnoreKey := common.IgnoreMetadataKey
			common.IgnoreMetadataKey = tt.ignoreMetadataKey
			defer func() {
				common.IgnoreMetadataKey = originalIgnoreKey
			}()

			result := IsStringMapSubset(tt.masterMap, tt.subsetMap)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestStringArrayContains(t *testing.T) {
	tests := []struct {
		name     string
		array    []string
		search   string
		expected bool
	}{
		{
			name:     "String found in array",
			array:    []string{"apple", "banana", "cherry"},
			search:   "banana",
			expected: true,
		},
		{
			name:     "String not found in array",
			array:    []string{"apple", "banana", "cherry"},
			search:   "grape",
			expected: false,
		},
		{
			name:     "Empty array",
			array:    []string{},
			search:   "apple",
			expected: false,
		},
		{
			name:     "Empty string search in non-empty array",
			array:    []string{"apple", "banana", "cherry"},
			search:   "",
			expected: false,
		},
		{
			name:     "Empty string search in array with empty string",
			array:    []string{"apple", "", "cherry"},
			search:   "",
			expected: true,
		},
		{
			name:     "Case sensitive search",
			array:    []string{"Apple", "Banana", "Cherry"},
			search:   "apple",
			expected: false,
		},
		{
			name:     "Array with duplicates",
			array:    []string{"apple", "banana", "apple", "cherry"},
			search:   "apple",
			expected: true,
		},
		{
			name:     "Single element array - found",
			array:    []string{"apple"},
			search:   "apple",
			expected: true,
		},
		{
			name:     "Single element array - not found",
			array:    []string{"apple"},
			search:   "banana",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringArrayContains(tt.array, tt.search)
			assert.Equal(t, tt.expected, result, "StringArrayContains(%v, %q) should return %v", tt.array, tt.search, tt.expected)
		})
	}
}

func TestConvertYamlToK8s(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		expectError bool
		description string
	}{
		{
			name: "Valid Deployment YAML",
			yamlContent: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
      - name: test-container
        image: nginx:latest`,
			expectError: false,
			description: "Should successfully parse valid Deployment YAML",
		},
		{
			name: "Valid Service YAML",
			yamlContent: `apiVersion: v1
kind: Service
metadata:
  name: test-service
  namespace: default
spec:
  selector:
    app: test
  ports:
  - port: 80
    targetPort: 8080`,
			expectError: false,
			description: "Should successfully parse valid Service YAML",
		},
		{
			name: "Valid ConfigMap YAML",
			yamlContent: `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-configmap
  namespace: default
data:
  key1: value1
  key2: value2`,
			expectError: false,
			description: "Should successfully parse valid ConfigMap YAML",
		},
		{
			name:        "Invalid YAML syntax",
			yamlContent: `invalid: yaml: content: [`,
			expectError: true,
			description: "Should return error for invalid YAML syntax",
		},
		{
			name:        "Empty YAML",
			yamlContent: "",
			expectError: true,
			description: "Should return error for empty YAML",
		},
		{
			name: "Non-Kubernetes YAML",
			yamlContent: `name: test
description: This is not a Kubernetes resource
data:
  key: value`,
			expectError: true,
			description: "Should return error for non-Kubernetes YAML",
		},
		{
			name: "YAML with unknown fields",
			yamlContent: `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-configmap
  namespace: default
unknownField: value
data:
  key1: value1`,
			expectError: false,
			description: "Should handle YAML with unknown fields gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, err := ConvertYamlToK8s([]byte(tt.yamlContent))

			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, obj, "Object should be nil when error occurs")
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, obj, "Object should not be nil when successful")

				// Verify the object has expected properties for known types
				switch tt.name {
				case "Valid Deployment YAML":
					deployment, ok := obj.(*appsv1.Deployment)
					assert.True(t, ok, "Object should be a Deployment")
					assert.Equal(t, "test-deployment", deployment.Name)
					assert.Equal(t, "default", deployment.Namespace)
				case "Valid Service YAML":
					service, ok := obj.(*corev1.Service)
					assert.True(t, ok, "Object should be a Service")
					assert.Equal(t, "test-service", service.Name)
					assert.Equal(t, "default", service.Namespace)
				case "Valid ConfigMap YAML":
					configMap, ok := obj.(*corev1.ConfigMap)
					assert.True(t, ok, "Object should be a ConfigMap")
					assert.Equal(t, "test-configmap", configMap.Name)
					assert.Equal(t, "default", configMap.Namespace)
				}
			}
		})
	}
}

func TestReadK8sFromFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := ioutil.TempDir("", "utils_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to temp directory for testing
	originalDir, err := os.Getwd()
	assert.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	// Valid YAML content for testing
	validDeploymentYAML := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
      - name: test-container
        image: nginx:latest`

	invalidYAML := `invalid: yaml: content: [`

	tests := []struct {
		name        string
		setupFile   func() string
		expectError bool
		description string
	}{
		{
			name: "Valid YAML file",
			setupFile: func() string {
				filename := "valid-deployment.yaml"
				err := ioutil.WriteFile(filename, []byte(validDeploymentYAML), 0644)
				assert.NoError(t, err)
				return filename
			},
			expectError: false,
			description: "Should successfully read and parse valid YAML file",
		},
		{
			name: "Non-existent file",
			setupFile: func() string {
				return "non-existent-file.yaml"
			},
			expectError: true,
			description: "Should return error for non-existent file",
		},
		{
			name: "Invalid YAML file",
			setupFile: func() string {
				filename := "invalid.yaml"
				err := ioutil.WriteFile(filename, []byte(invalidYAML), 0644)
				assert.NoError(t, err)
				return filename
			},
			expectError: true,
			description: "Should return error for invalid YAML content",
		},
		{
			name: "Empty file",
			setupFile: func() string {
				filename := "empty.yaml"
				err := ioutil.WriteFile(filename, []byte(""), 0644)
				assert.NoError(t, err)
				return filename
			},
			expectError: true,
			description: "Should return error for empty file",
		},
		{
			name: "File with no read permissions",
			setupFile: func() string {
				filename := "no-read-permission.yaml"
				err := ioutil.WriteFile(filename, []byte(validDeploymentYAML), 0000)
				assert.NoError(t, err)
				return filename
			},
			expectError: true,
			description: "Should return error for file with no read permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := tt.setupFile()

			obj, err := ReadK8sFromFile(filename)

			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, obj, "Object should be nil when error occurs")
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, obj, "Object should not be nil when successful")

				// For valid deployment, verify it's parsed correctly
				if tt.name == "Valid YAML file" {
					deployment, ok := obj.(*appsv1.Deployment)
					assert.True(t, ok, "Object should be a Deployment")
					assert.Equal(t, "test-deployment", deployment.Name)
					assert.Equal(t, "default", deployment.Namespace)
				}
			}
		})
	}
}

func TestSetup(t *testing.T) {
	// Save original KubeconfigPath
	originalKubeconfigPath := common.KubeconfigPath
	defer func() {
		common.KubeconfigPath = originalKubeconfigPath
	}()

	tests := []struct {
		name           string
		kubeconfigPath string
		expectError    bool
		description    string
		skipActualCall bool
	}{
		{
			name:           "Empty KubeconfigPath - should call in-cluster config",
			kubeconfigPath: "",
			expectError:    true, // Will fail in test environment as we're not in cluster
			description:    "Should attempt in-cluster config when KubeconfigPath is empty",
			skipActualCall: false,
		},
		{
			name:           "Non-empty KubeconfigPath - should call external config",
			kubeconfigPath: "/path/to/kubeconfig",
			expectError:    true, // Will fail as path doesn't exist
			description:    "Should attempt external config when KubeconfigPath is set",
			skipActualCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the KubeconfigPath for this test
			common.KubeconfigPath = tt.kubeconfigPath

			config, clientset, err := Setup()

			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, config, "Config should be nil when error occurs")
				assert.Nil(t, clientset, "Clientset should be nil when error occurs")
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, config, "Config should not be nil when successful")
				assert.NotNil(t, clientset, "Clientset should not be nil when successful")
			}
		})
	}
}

func TestSetupArgoRollout(t *testing.T) {
	// Save original KubeconfigPath
	originalKubeconfigPath := common.KubeconfigPath
	defer func() {
		common.KubeconfigPath = originalKubeconfigPath
	}()

	tests := []struct {
		name           string
		kubeconfigPath string
		expectError    bool
		description    string
	}{
		{
			name:           "Empty KubeconfigPath - should call in-cluster argo config",
			kubeconfigPath: "",
			expectError:    true, // Will fail in test environment as we're not in cluster
			description:    "Should attempt in-cluster argo config when KubeconfigPath is empty",
		},
		{
			name:           "Non-empty KubeconfigPath - should call external argo config",
			kubeconfigPath: "/path/to/kubeconfig",
			expectError:    true, // Will fail as path doesn't exist
			description:    "Should attempt external argo config when KubeconfigPath is set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the KubeconfigPath for this test
			common.KubeconfigPath = tt.kubeconfigPath

			config, clientset, err := SetupArgoRollout()

			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, config, "Config should be nil when error occurs")
				assert.Nil(t, clientset, "Clientset should be nil when error occurs")
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, config, "Config should not be nil when successful")
				assert.NotNil(t, clientset, "Clientset should not be nil when successful")
			}
		})
	}
}
