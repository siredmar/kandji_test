package v20190529

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	types "github.com/grid-x/ds-api-types"
)

// DeviceCleanupConfig as exposed by this API version
type DeviceCleanupConfig struct {
	Metadata types.Metadata            `json:"metadata,omitempty"`
	Spec     DeviceCleanupConfigSpec   `json:"spec"`
	Status   DeviceCleanupConfigStatus `json:"status"`
}

// DeviceCleanupConfigSpec defines the desired state of DeviceCleanupConfig
type DeviceCleanupConfigSpec struct {
	// DeviceID is the ID of the device this config should applied to
	DeviceID string              `json:"deviceID"`
	Docker   DockerCleanupConfig `json:"docker"`
}

// DockerCleanupConfig defines the desired state of CleanupConfig
type DockerCleanupConfig struct {
	Container DockerContainerCleanupConfig `json:"container"`
	Images    DockerImagesCleanupConfig    `json:"images"`
}

// DockerImagesCleanupConfig defines the desired state of image CleanupConfig
type DockerImagesCleanupConfig struct {
	// Can be defined naturally eg. 5d / 1w / 2m. Images older than the defined retention period will be cleaned up
	// with respect to the number of minimum images
	RetentionPeriod string `json:"retentionPeriod"`
	// Defines the number of n last images per repository which should not be affected by the cleanup even if they've exceeded the retention time
	MinimumImages int `json:"minImages"`
}

// DockerContainerCleanupConfig defines the desired state of container CleanupConfig
type DockerContainerCleanupConfig struct {
	// Can be defined naturally eg. 5d / 1w / 2m. Existed container older than the defined retention period will be cleaned up
	RetentionPeriod string `json:"retentionPeriod"`
}

// DeviceCleanupConfigStatus defines the observed state of DeviceCleanupConfig
type DeviceCleanupConfigStatus struct {
	AppliedAt *metav1.Time `json:"appliedAt,omitempty"`
}
