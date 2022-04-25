/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FlippyConfigSpec defines the desired state of FlippyConfig
type FlippyConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//List of allowed docker images
	ImageList []string `json:"ImageList,omitempty"`

	//List of precondition before rotating any pod
	Preconditions []FlippyCondition `json:"Preconditions,omitempty"`

	//List of conditions on which pods get filter
	ProcessFilter ProcessFilter `json:"ProcessFilter,omitempty"`

	RestartObjects []RestartObject `json:"RestartObjects,omitempty"`

	PostFilterRestarts []FlippyCondition `json:"PostFilterRestarts,omitempty"`
}

type RestartObject struct {

	//Type of object to be restarted
	Type string `json:"Type,omitempty"`

	//Retry configuration for status check
	StatusCheckConfig StatusCheckConfig `json:"StatusCheckConfig,omitempty"`
}

type ProcessFilter struct {
	//Metdata Pod Label Filter
	PodLabels map[string]string `json:"PodLabels,omitempty"`

	//Metdata Namespace Label Filter
	NamespaceLabels map[string]string `json:"Labels,omitempty"`

	//List of annotation on object eg. Deployment ArgoRollouts
	Annotations map[string]string `json:"Annotations,omitempty"`

	//Container names
	Containers []string `json:"Containers,omitempty"`

	PreProcessRestart K8S `json:"PreProcessRestart,omitempty"`
}

type K8S struct {
	//Type of object. E.g Deployment
	Type string `json:"Type,omitempty"`

	//Name of the object
	Name string `json:"Name,omitempty"`

	//Namespace it belongs
	Namespace string `json:"Namespace,omitempty"`
}

type FlippyCondition struct {

	//Kubernetes Object
	K8S K8S `json:"K8S,omitempty"`

	//Status to be verified
	Status string `json:"Status,omitempty"`

	//Retry configuration for status check
	StatusCheckConfig StatusCheckConfig `json:"StatusCheckConfig,omitempty"`
}

type StatusCheckConfig struct {

	//CheckStatus after
	CheckStatus bool `json:"CheckStatus,omitempty"`

	//Retries to check status
	MaxRetry int `json:"MaxRetry,omitempty"`

	//Retry duration in seconds
	RetryDuration int `json:"RetryDuration,omitempty"`
}

// FlippyConfigStatus defines the observed state of FlippyConfig
type FlippyConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// FlippyConfig is the Schema for the FlippyConfig API
type FlippyConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FlippyConfigSpec   `json:"spec,omitempty"`
	Status FlippyConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FlippyConfigList contains a list of FlippyConfig
type FlippyConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FlippyConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FlippyConfig{}, &FlippyConfigList{})
}
