package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DeviceDeploymentSpec defines the desired state of DeviceDeployment
type DeviceDeploymentSpec struct {
	// The App this deployment belongs to. This is required to decide which
	// deployment to use if there are multiple deployments of the same app
	App string `json:"app"`
	// Selector defines to which gridBoxes this deployment can be applied
	Selector Selector `json:"selector"`
	// Template defines the template of the pod that will be created
	Template PodTemplate `json:"template"`
}

// DeviceDeploymentStatus defines the observed state of DeviceDeployment
type DeviceDeploymentStatus struct {
	LastUpdatedAt string `json:"LastUpdatedAt,omitempty"`
}

// Selector defines which gridBoxes are selected by this
// workload.
// It allows to either select via labels or via deviceID.
// If both are set all gridBoxes will be selected that match one of these.
type Selector struct {
	// Match by label. If all conditions (i.e. matching labels) are true
	// the gridBox will be selected
	MatchByLabels map[string]string `json:"matchByLabels"`
	// Match by device ID directly. Will always overwrite the match by
	// labels value
	// +optional
	MatchByDeviceID *string `json:"matchByDeviceID,omitempty"`
}

// PodTemplate is the template for the spec of the pod that is going to be
// created
type PodTemplate struct {
	Spec corev1beta1.PodConfig `json:"spec"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DeviceDeployment is the Schema for the devicedeployments API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type DeviceDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeviceDeploymentSpec   `json:"spec,omitempty"`
	Status DeviceDeploymentStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DeviceDeploymentList contains a list of DeviceDeployment
type DeviceDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DeviceDeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DeviceDeployment{}, &DeviceDeploymentList{})
}
