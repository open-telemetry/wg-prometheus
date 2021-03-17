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

// CollectorLoadGeneratorSpec defines the desired state of CollectorLoadGenerator
type CollectorLoadGeneratorSpec struct {
	// Replicas specify the exact number of sample applications that generate load
	// TODO: validate minimal is 1, there should be some kubebuilder marker
	Replicas int32 `json:"replicas,omitempty"`
}

// CollectorLoadGeneratorStatus defines the observed state of CollectorLoadGenerator
type CollectorLoadGeneratorStatus struct {
	Replicas int32  `json:"replicas,omitempty"`
	Selector string `json:"selector"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CollectorLoadGenerator is the Schema for the collectorloadgenerators API.
type CollectorLoadGenerator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CollectorLoadGeneratorSpec   `json:"spec,omitempty"`
	Status CollectorLoadGeneratorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.selector

// CollectorLoadGeneratorList contains a list of CollectorLoadGenerator
type CollectorLoadGeneratorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CollectorLoadGenerator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CollectorLoadGenerator{}, &CollectorLoadGeneratorList{})
}
