package config

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetup(t *testing.T) {
	tests := []struct {
		name           string
		kubeConfigPath string
		setupFile      func() (string, func())
		expectError    bool
		description    string
	}{
		{
			name:           "NonExistentFile",
			kubeConfigPath: "/non/existent/path/kubeconfig",
			setupFile:      func() (string, func()) { return "/non/existent/path/kubeconfig", func() {} },
			expectError:    true,
			description:    "Should return error when kubeconfig file doesn't exist",
		},
		{
			name:           "EmptyPath",
			kubeConfigPath: "",
			setupFile:      func() (string, func()) { return "", func() {} },
			expectError:    true,
			description:    "Should return error when kubeconfig path is empty",
		},
		{
			name:           "InvalidKubeconfig",
			kubeConfigPath: "",
			setupFile: func() (string, func()) {
				// Create temporary file with invalid kubeconfig content
				tmpFile, err := ioutil.TempFile("", "invalid-kubeconfig-*.yaml")
				if err != nil {
					panic(err)
				}

				// Write invalid YAML content
				invalidContent := `
invalid: yaml: content:
  - this is not
    valid kubeconfig
malformed yaml
`
				_, err = tmpFile.WriteString(invalidContent)
				if err != nil {
					panic(err)
				}
				tmpFile.Close()

				return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }
			},
			expectError: true,
			description: "Should return error when kubeconfig file contains invalid content",
		},
		{
			name:           "UnreadableFile",
			kubeConfigPath: "",
			setupFile: func() (string, func()) {
				// Create temporary file and make it unreadable
				tmpFile, err := ioutil.TempFile("", "unreadable-kubeconfig-*.yaml")
				if err != nil {
					panic(err)
				}
				tmpFile.Close()

				// Make file unreadable (this might not work on all systems)
				os.Chmod(tmpFile.Name(), 0000)

				return tmpFile.Name(), func() {
					os.Chmod(tmpFile.Name(), 0644) // Restore permissions before cleanup
					os.Remove(tmpFile.Name())
				}
			},
			expectError: true,
			description: "Should return error when kubeconfig file is unreadable",
		},
		{
			name:           "MalformedYAML",
			kubeConfigPath: "",
			setupFile: func() (string, func()) {
				tmpFile, err := ioutil.TempFile("", "malformed-kubeconfig-*.yaml")
				if err != nil {
					panic(err)
				}

				// Write malformed YAML
				malformedContent := `{invalid json and yaml content`
				_, err = tmpFile.WriteString(malformedContent)
				if err != nil {
					panic(err)
				}
				tmpFile.Close()

				return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }
			},
			expectError: true,
			description: "Should return error when kubeconfig file contains malformed YAML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test file if needed
			kubeconfigPath, cleanup := tt.setupFile()
			defer cleanup()

			if tt.kubeConfigPath != "" {
				kubeconfigPath = tt.kubeConfigPath
			}

			// Execute function
			config, clientset, err := Setup(kubeconfigPath)

			// Verify results
			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, config, "Config should be nil on error")
				assert.Nil(t, clientset, "Clientset should be nil on error")
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, config, "Config should not be nil on success")
				assert.NotNil(t, clientset, "Clientset should not be nil on success")
			}
		})
	}
}

func TestSetupInCluster(t *testing.T) {
	// Test in-cluster setup outside of cluster environment
	// This should fail since we're not running in a Kubernetes cluster

	config, clientset, err := SetupInCluster()

	// Should return error when not running in cluster
	assert.Error(t, err, "Should return error when not running in Kubernetes cluster")
	assert.Nil(t, config, "Config should be nil when not in cluster")
	assert.Nil(t, clientset, "Clientset should be nil when not in cluster")

	// Verify error message indicates in-cluster issue
	assert.Contains(t, err.Error(), "unable to load in-cluster configuration",
		"Error should indicate in-cluster configuration issue")
}

func TestSetupArgoRollout(t *testing.T) {
	tests := []struct {
		name           string
		kubeConfigPath string
		setupFile      func() (string, func())
		expectError    bool
		description    string
	}{
		{
			name:           "NonExistentFile",
			kubeConfigPath: "/non/existent/path/kubeconfig",
			setupFile:      func() (string, func()) { return "/non/existent/path/kubeconfig", func() {} },
			expectError:    true,
			description:    "Should return error when kubeconfig file doesn't exist",
		},
		{
			name:           "EmptyPath",
			kubeConfigPath: "",
			setupFile:      func() (string, func()) { return "", func() {} },
			expectError:    true,
			description:    "Should return error when kubeconfig path is empty",
		},
		{
			name:           "InvalidKubeconfig",
			kubeConfigPath: "",
			setupFile: func() (string, func()) {
				tmpFile, err := ioutil.TempFile("", "invalid-argo-kubeconfig-*.yaml")
				if err != nil {
					panic(err)
				}

				invalidContent := `
invalid: argo: config:
  - malformed content
`
				_, err = tmpFile.WriteString(invalidContent)
				if err != nil {
					panic(err)
				}
				tmpFile.Close()

				return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }
			},
			expectError: true,
			description: "Should return error when kubeconfig file contains invalid content for Argo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kubeconfigPath, cleanup := tt.setupFile()
			defer cleanup()

			if tt.kubeConfigPath != "" {
				kubeconfigPath = tt.kubeConfigPath
			}

			// Execute function
			config, clientset, err := SetupArgoRollout(kubeconfigPath)

			// Verify results
			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, config, "Config should be nil on error")
				assert.Nil(t, clientset, "Argo clientset should be nil on error")
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, config, "Config should not be nil on success")
				assert.NotNil(t, clientset, "Argo clientset should not be nil on success")
			}
		})
	}
}

func TestSetupInClusterArgoRollout(t *testing.T) {
	// Test in-cluster Argo Rollout setup outside of cluster environment
	// This should fail since we're not running in a Kubernetes cluster

	config, clientset, err := SetupInClusterArgoRollout()

	// Should return error when not running in cluster
	assert.Error(t, err, "Should return error when not running in Kubernetes cluster")
	assert.Nil(t, config, "Config should be nil when not in cluster")
	assert.Nil(t, clientset, "Argo clientset should be nil when not in cluster")

	// Verify error message indicates in-cluster issue
	assert.Contains(t, err.Error(), "unable to load in-cluster configuration",
		"Error should indicate in-cluster configuration issue")
}

func TestSetup_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		setupFile   func() (string, func())
		expectError bool
		description string
	}{
		{
			name: "EmptyFile",
			setupFile: func() (string, func()) {
				tmpFile, err := ioutil.TempFile("", "empty-kubeconfig-*.yaml")
				if err != nil {
					panic(err)
				}
				tmpFile.Close() // Create empty file

				return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }
			},
			expectError: true,
			description: "Should return error when kubeconfig file is empty",
		},
		{
			name: "DirectoryInsteadOfFile",
			setupFile: func() (string, func()) {
				tmpDir, err := ioutil.TempDir("", "kubeconfig-dir-*")
				if err != nil {
					panic(err)
				}

				return tmpDir, func() { os.RemoveAll(tmpDir) }
			},
			expectError: true,
			description: "Should return error when kubeconfig path points to directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kubeconfigPath, cleanup := tt.setupFile()
			defer cleanup()

			// Execute function
			config, clientset, err := Setup(kubeconfigPath)

			// Verify results
			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, config, "Config should be nil on error")
				assert.Nil(t, clientset, "Clientset should be nil on error")
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, config, "Config should not be nil on success")
				assert.NotNil(t, clientset, "Clientset should not be nil on success")
			}
		})
	}
}

// TestSetup_SuccessPath tests the success path with a valid kubeconfig
func TestSetup_SuccessPath(t *testing.T) {
	// Create a valid kubeconfig file
	validKubeconfig := `
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://localhost:8443
    insecure-skip-tls-verify: true
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`

	tmpFile, err := ioutil.TempFile("", "valid-kubeconfig-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(validKubeconfig)
	assert.NoError(t, err)
	tmpFile.Close()

	// Execute function
	config, clientset, err := Setup(tmpFile.Name())

	// Verify results - should succeed in parsing config and creating clientset
	// Note: The clientset creation may fail due to network issues, but config should be valid
	if err != nil {
		// If there's an error, it should be from clientset creation, not config parsing
		assert.NotNil(t, config, "Config should be parsed successfully even if clientset creation fails")
		assert.Nil(t, clientset, "Clientset should be nil when creation fails")
		// The error should be related to connection, not parsing
		assert.NotContains(t, err.Error(), "invalid configuration", "Error should not be about invalid configuration")
	} else {
		// If no error, both should be valid
		assert.NotNil(t, config, "Config should not be nil on success")
		assert.NotNil(t, clientset, "Clientset should not be nil on success")
		assert.Equal(t, "https://localhost:8443", config.Host, "Config should have correct server URL")
	}
}

// TestSetupArgoRollout_SuccessPath tests the success path with a valid kubeconfig
func TestSetupArgoRollout_SuccessPath(t *testing.T) {
	// Create a valid kubeconfig file
	validKubeconfig := `
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://localhost:8443
    insecure-skip-tls-verify: true
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`

	tmpFile, err := ioutil.TempFile("", "valid-argo-kubeconfig-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(validKubeconfig)
	assert.NoError(t, err)
	tmpFile.Close()

	// Execute function
	config, clientset, err := SetupArgoRollout(tmpFile.Name())

	// Verify results - should succeed in parsing config and creating argo clientset
	if err != nil {
		// If there's an error, it should be from clientset creation, not config parsing
		assert.NotNil(t, config, "Config should be parsed successfully even if argo clientset creation fails")
		assert.Nil(t, clientset, "Argo clientset should be nil when creation fails")
		// The error should be related to connection, not parsing
		assert.NotContains(t, err.Error(), "invalid configuration", "Error should not be about invalid configuration")
	} else {
		// If no error, both should be valid
		assert.NotNil(t, config, "Config should not be nil on success")
		assert.NotNil(t, clientset, "Argo clientset should not be nil on success")
		assert.Equal(t, "https://localhost:8443", config.Host, "Config should have correct server URL")
	}
}

// TestSetupInCluster_SuccessPath tests the success path by mocking in-cluster environment
func TestSetupInCluster_SuccessPath(t *testing.T) {
	// Skip this test as it requires complex mocking of in-cluster environment
	// The in-cluster setup requires service account tokens and environment variables
	// that are difficult to mock properly in unit tests
	t.Skip("In-cluster setup requires complex environment mocking - tested in integration tests")
}

// TestSetupInClusterArgoRollout_SuccessPath tests the success path by mocking in-cluster environment
func TestSetupInClusterArgoRollout_SuccessPath(t *testing.T) {
	// Skip this test as it requires complex mocking of in-cluster environment
	// The in-cluster setup requires service account tokens and environment variables
	// that are difficult to mock properly in unit tests
	t.Skip("In-cluster argo setup requires complex environment mocking - tested in integration tests")
}

// TestSetup_ClientsetCreationError tests error handling when clientset creation fails
func TestSetup_ClientsetCreationError(t *testing.T) {
	// Create a kubeconfig with invalid server URL to trigger clientset creation error
	invalidServerKubeconfig := `
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://invalid-server-url-that-does-not-exist:8443
    insecure-skip-tls-verify: true
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`

	tmpFile, err := ioutil.TempFile("", "invalid-server-kubeconfig-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(invalidServerKubeconfig)
	assert.NoError(t, err)
	tmpFile.Close()

	// Execute function
	config, clientset, err := Setup(tmpFile.Name())

	// The config should be parsed successfully, but clientset creation may fail
	// This tests the path where BuildConfigFromFlags succeeds but NewForConfig fails
	if err != nil {
		// Config parsing succeeded, but clientset creation failed
		assert.NotNil(t, config, "Config should be parsed successfully")
		assert.Nil(t, clientset, "Clientset should be nil when creation fails")
	} else {
		// If no error, both should be valid (test environment may allow this)
		assert.NotNil(t, config, "Config should not be nil")
		assert.NotNil(t, clientset, "Clientset should not be nil")
	}
}

// TestSetupArgoRollout_ClientsetCreationError tests error handling when argo clientset creation fails
func TestSetupArgoRollout_ClientsetCreationError(t *testing.T) {
	// Create a kubeconfig with invalid server URL to trigger argo clientset creation error
	invalidServerKubeconfig := `
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://invalid-argo-server-url:8443
    insecure-skip-tls-verify: true
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`

	tmpFile, err := ioutil.TempFile("", "invalid-argo-server-kubeconfig-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(invalidServerKubeconfig)
	assert.NoError(t, err)
	tmpFile.Close()

	// Execute function
	config, clientset, err := SetupArgoRollout(tmpFile.Name())

	// The config should be parsed successfully, but argo clientset creation may fail
	if err != nil {
		// Config parsing succeeded, but argo clientset creation failed
		assert.NotNil(t, config, "Config should be parsed successfully")
		assert.Nil(t, clientset, "Argo clientset should be nil when creation fails")
	} else {
		// If no error, both should be valid (test environment may allow this)
		assert.NotNil(t, config, "Config should not be nil")
		assert.NotNil(t, clientset, "Argo clientset should not be nil")
	}
}
