package v20190817

import (
	types "github.com/grid-x/ds-api-types"
	v20190817Pod "github.com/grid-x/ds-api-types/management/2019-08-17/pod"
)

// Deployment represents the deployment type as exported by this API version
type Deployment struct {
	Metadata types.Metadata         `json:"metadata,omitempty"`
	Spec     DeviceDeploymentSpec   `json:"spec"`
	Status   DeviceDeploymentStatus `json:"status"`
}

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
	Spec v20190817Pod.PodConfig `json:"spec"`
}
