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

// CollectorLoadBalancerSpec defines the desired state of CollectorLoadBalancer
type CollectorLoadBalancerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Replicas specify the exact number of replicas.
	// When set to 0, auto scale is enabled with a default max set to 10.
	// FIXME: the line above is lying ... it just panic if you don't enable scale directly
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// When scale is enabled, Replicas is ignored
	// +optional
	Scale ScalerConfig `json:"scale,omitempty"`

	// CollectorImage specify which collector image to use
	CollectorImage string `json:"collectorImage,omitempty"`
	// CollectorConfig is (the entire) otel collector config.
	CollectorConfig string `json:"collectorConfig,omitempty"`
}

// CollectorLoadBalancerStatus defines the observed state of CollectorLoadBalancer
type CollectorLoadBalancerStatus struct {
	Replicas int32 `json:"replicas,omitempty"`
	// TODO: what set it to the right string form
	// TODO: so what does HPA look at? the replicas in status? and then update replicas in spec?
	// https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/ seems deployment, sts etc. all works
	// https://book.kubebuilder.io/reference/generating-crd.html#scale
	// https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#scale-subresource
	// https://medium.com/@thescott111/autoscaling-kubernetes-custom-resource-using-the-hpa-957d00bb7993
	Selector string `json:"selector"` // this must be the string form of the selector
}

// Scale related config
type ScalerConfig struct {
	Enabled                     bool `json:"enabled,omitempty"`
	ExpectedTargetsPerCollector int  `json:"expected_targets_per_collector,omitempty"`
	MinReplicas                 int  `json:"min_replicas,omitempty"`
	MaxReplicas                 int  `json:"max_replicas,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.selector

// CollectorLoadBalancer is the Schema for the collectorloadbalancers API
type CollectorLoadBalancer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CollectorLoadBalancerSpec   `json:"spec,omitempty"`
	Status CollectorLoadBalancerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CollectorLoadBalancerList contains a list of CollectorLoadBalancer
type CollectorLoadBalancerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CollectorLoadBalancer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CollectorLoadBalancer{}, &CollectorLoadBalancerList{})
}
