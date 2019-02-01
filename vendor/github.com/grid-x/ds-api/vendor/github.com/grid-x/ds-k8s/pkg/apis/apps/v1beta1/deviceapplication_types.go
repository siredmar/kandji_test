/*
 * Author: Joel Hermanns <j.hermanns@gridx.ai>
 */

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeviceApplicationSpec defines the desired state of DeviceApplication
type DeviceApplicationSpec struct {
	Name string `json:"name"`
}

// DeviceApplicationStatus defines the observed state of DeviceApplication
// NOTE: An application has no status right
type DeviceApplicationStatus struct{}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DeviceApplication is the Schema for the deviceapplications API
// +k8s:openapi-gen=true
type DeviceApplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeviceApplicationSpec   `json:"spec,omitempty"`
	Status DeviceApplicationStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DeviceApplicationList contains a list of DeviceApplication
type DeviceApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DeviceApplication `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DeviceApplication{}, &DeviceApplicationList{})
}
