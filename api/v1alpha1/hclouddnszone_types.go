/*
Copyright 2025.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HcloudDnsZoneSpec defines the desired state of HcloudDnsZone
type HcloudDnsZoneSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// +required
	Name string `json:"name"`

	// +optional
	// +kubebuilder:validation:Enum=PRIMARY;SECONDARY
	Mode string `json:"mode,omitempty"`

	// +optional
	TTL *int `json:"ttl,omitempty"`

	// +optional
	Labels map[string]string `json:"labels,omitempty"`
}

// HcloudDnsZoneStatus defines the observed state of HcloudDnsZone.
type HcloudDnsZoneStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties
	ZoneId int `json:"zoneId,omitempty"`

	Mode string `json:"mode,omitempty"`

	TTL int `json:"ttl,omitempty"`

	Labels map[string]string `json:"labels,omitempty"`

	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// conditions represent the current state of the HcloudDnsZone resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// Standard condition types include:
	// - "Available": the resource is fully functional
	// - "Progressing": the resource is being created or updated
	// - "Degraded": the resource failed to reach or maintain its desired state
	//
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="ZoneId",type=integer,JSONPath=`.status.zoneId`,description="Hetzner Cloud DNS Zone ID"
// +kubebuilder:printcolumn:name="ProvisioningState",type=string,JSONPath=`.status.conditions[?(@.type=="Available")].reason`,description="Provisioning state of the network"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`,description="Age of the resource"

// HcloudDnsZone is the Schema for the hclouddnszones API
type HcloudDnsZone struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of HcloudDnsZone
	// +required
	Spec HcloudDnsZoneSpec `json:"spec"`

	// status defines the observed state of HcloudDnsZone
	// +optional
	Status HcloudDnsZoneStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// HcloudDnsZoneList contains a list of HcloudDnsZone
type HcloudDnsZoneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []HcloudDnsZone `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HcloudDnsZone{}, &HcloudDnsZoneList{})
}
