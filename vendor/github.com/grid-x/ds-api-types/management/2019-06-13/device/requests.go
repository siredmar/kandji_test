package v20190613

import (
	types "github.com/grid-x/ds-api-types"
	v20190817Pod "github.com/grid-x/ds-api-types/management/2019-08-17/pod"
)

// CreateRequest represents the request type
// Endpoint: POST /devices
type CreateRequest struct {
	Metadata types.Metadata `json:"metadata"`
	Spec     DeviceSpec     `json:"spec"`
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

// ListPodsRequest represents the request type
// Endpoint: GET /devices/{deviceID}/pods
type ListPodsRequest struct{}

// ListPodsResponse represents the response type
type ListPodsResponse struct {
	Pods []*v20190817Pod.Pod `json:"pods"`
}
