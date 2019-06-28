package v20190529

import (
	types "github.com/grid-x/ds-api-types"
	v20181128Deployment "github.com/grid-x/ds-api-types/management/2018-11-28/deployments"
)

// CleanupConfig as exposed by this API version
type CleanupConfig struct {
	Metadata types.Metadata      `json:"metadata,omitempty"`
	Spec     CleanupConfigSpec   `json:"spec"`
	Status   CleanupConfigStatus `json:"status"`
}

// CleanupConfigSpec defines the desired state of CleanupConfig
type CleanupConfigSpec struct {
	Docker   DockerCleanupConfig          `json:"docker"`
	Selector v20181128Deployment.Selector `json:"selector"`
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

// CleanupConfigStatus defines the observed state of CleanupConfig
type CleanupConfigStatus struct {
}
