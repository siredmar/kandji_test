package v20190613

import (
	types "github.com/grid-x/ds-api-types"
	v20181128Pod "github.com/grid-x/ds-api-types/management/2018-11-28/pod"
	v20190529DeviceCleanupConfig "github.com/grid-x/ds-api-types/management/2019-05-29/devicecleanupconfigs"
	v20190609DeviceDockerConfig "github.com/grid-x/ds-api-types/management/2019-06-09/devicedockerconfigs"
)

// CreateRequest represents the request type
// Endpoint: POST /devices
type CreateRequest struct {
	Metadata *types.CreateMetadata `json:"metadata,omitempty"`
	Spec     DeviceSpec            `json:"spec"`
}

// CreateResponse represents the response type
type CreateResponse struct {
	*Device
}

// ListRequest represents the request type
// Endpoint: GET /devices
type ListRequest struct{}

// ListResponse represents the response type
type ListResponse struct {
	Devices []*Device `json:"devices"`
}

// ListBySerialNumberRequest represents the request type
// Endpoint: GET /devices?filter=serialnumber:*314-P-X
type ListBySerialNumberRequest struct{}

// ListBySerialNumberResponse represents the response type
type ListBySerialNumberResponse struct {
	Devices []*Device `json:"devices"`
}

// ListByProductionNumberRequest represents the request type
// Endpoint: GET /devices?filter=productionnumber:120
type ListByProductionNumberRequest struct{}

// ListByProductionNumberResponse represents the response type
type ListByProductionNumberResponse struct {
	Devices []*Device `json:"devices"`
}

// GetRequest represents the request type
// Endpoint: GET /devices/{deviceID}
type GetRequest struct{}

// GetResponse represents the response type
type GetResponse struct {
	*Device
}

// UpdateRequest represents the request type
// Endpoint: PATCH /devices/{deviceID}
type UpdateRequest struct {
	Metadata *types.UpdateMetadata `json:"metadata,omitempty"`
	Spec     UpdateSpec            `json:"spec,omitempty"`
}

// UpdateResponse represents the response type
type UpdateResponse struct {
	*Device
}

// DeleteRequest represents the request type
// Endpoint: DELETE /devices/{deviceID}
type DeleteRequest struct{}

// DeleteResponse represents the response type
type DeleteResponse struct{}

// ListDeviceDockerConfigsRequest represents the request type
// Endpoint: GET /devices/{deviceID}/devicedockerconfigs
type ListDeviceDockerConfigsRequest struct{}

// ListDeviceDockerConfigsResponse represents the response type
type ListDeviceDockerConfigsResponse struct {
	DeviceDockerConfigs []*v20190609DeviceDockerConfig.DeviceDockerConfig `json:"deviceDockerConfigs"`
}

// ListDeviceCleanupConfigsRequest represents the request type
// Endpoint: GET /devices/{deviceID}/devicecleanupconfig
type ListDeviceCleanupConfigsRequest struct{}

// ListDeviceCleanupConfigsResponse represents the response type
type ListDeviceCleanupConfigsResponse struct {
	DeviceCleanupConfigs []*v20190529DeviceCleanupConfig.DeviceCleanupConfig `json:"deviceCleanupConfigs"`
}

// ListPodsRequest represents the request type
// Endpoint: GET /devices/{deviceID}/pods
type ListPodsRequest struct{}

// ListPodsResponse represents the response type
type ListPodsResponse struct {
	Pods []*v20181128Pod.Pod `json:"pods"`
}
