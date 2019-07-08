package v20190613

import (
	types "github.com/grid-x/ds-api-types"
	v20190529DeviceCleanupConfig "github.com/grid-x/ds-api-types/management/2019-05-29/devicecleanupconfigs"
	v20190609DeviceDockerConfig "github.com/grid-x/ds-api-types/management/2019-06-09/devicedockerconfigs"
)

// CreateRequest represents the request type
type CreateRequest struct {
	Metadata *types.CreateMetadata `json:"metadata,omitempty"`
	Spec     DeviceSpec            `json:"spec"`
}

// CreateResponse represents the response type
type CreateResponse struct {
	*Device
}

// ListResponse represents the response type
type ListResponse struct {
	Devices []*Device `json:"devices"`
}

// GetResponse represents the response type
type GetResponse struct {
	*Device
}

// UpdateRequest represents the request type
type UpdateRequest struct {
	Metadata *types.UpdateMetadata `json:"metadata,omitempty"`
	Spec     UpdateSpec            `json:"spec,omitempty"`
}

// UpdateResponse represents the response type
type UpdateResponse struct {
	*Device
}

// DeleteResponse represents the response type
type DeleteResponse struct{}

// ListDeviceDockerConfigsResponse represents the response type
type ListDeviceDockerConfigsResponse struct {
	DeviceDockerConfigs []*v20190609DeviceDockerConfig.DeviceDockerConfig `json:"deviceDockerConfigs"`
}

// ListDeviceCleanupConfigsResponse represents the response type
type ListDeviceCleanupConfigsResponse struct {
	DeviceCleanupConfigs []*v20190529DeviceCleanupConfig.DeviceCleanupConfig `json:"deviceCleanupConfigs"`
}
