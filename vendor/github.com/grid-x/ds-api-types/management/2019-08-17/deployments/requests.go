package v20190817

import (
	types "github.com/grid-x/ds-api-types"
	v20190613Device "github.com/grid-x/ds-api-types/management/2019-06-13/device"
)

// CreateRequest represents the request type
// Endpoint: POST /deployments
type CreateRequest struct {
	Spec *DeviceDeploymentSpec `json:"spec"`
}

// CreateResponse represents the response type
type CreateResponse struct {
	*Deployment
}

// ListRequest represents the request type
// Endpoint: GET /deployments
type ListRequest struct{}

// ListResponse represents the response type
type ListResponse struct {
	Deployments []*Deployment `json:"deployments"`
}

// GetRequest represents the request type
// Endpoint: GET /deployments/{deploymentID}
type GetRequest struct{}

// GetResponse represents the response type
type GetResponse struct {
	*Deployment
}

// UpdateRequest represents the request type
// Endpoint: PATCH /deployments/{deploymentID}
type UpdateRequest struct {
	Metadata types.UpdateMetadata  `json:"metadata"`
	Spec     *DeviceDeploymentSpec `json:"spec"`
}

// UpdateResponse represents the response type
type UpdateResponse struct {
	*Deployment
}

// DeleteRequest represents the request type
// Endpoint: DELETE /deployments/{deploymentID}
type DeleteRequest struct{}

// DeleteResponse represents the response type
type DeleteResponse struct{}

// ListDevicesRequest represents the request type
// Endpoint: GET /deployments/{deploymentID}/devices
type ListDevicesRequest struct{}

// ListDevicesResponse represents the response type
type ListDevicesResponse struct {
	Devices []*v20190613Device.Device `json:"devices"`
}
