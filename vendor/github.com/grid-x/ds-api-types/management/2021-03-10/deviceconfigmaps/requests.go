package v20210310

import (
	types "github.com/grid-x/ds-api-types"
)

// CreateRequest represents the request type
// Endpoint: POST /DeviceConfigMaps
type CreateRequest struct {
	Metadata types.Metadata       `json:"metadata"`
	Spec     *DeviceConfigMapSpec `json:"spec"`
}

// CreateResponse represents the response type
type CreateResponse struct {
	*DeviceConfigMap
}

// UpdateRequest represents the request type
// Endpoint: PATCH /DeviceConfigMaps/{DeviceConfigMapID}
type UpdateRequest struct {
	Metadata types.UpdateMetadata `json:"metadata"`
	Spec     *DeviceConfigMapSpec `json:"spec"`
}

// UpdateResponse represents the response type
type UpdateResponse struct {
	*DeviceConfigMap
}

// GetRequest represents the request type
// Endpoint: GET /DeviceConfigMaps/{DeviceConfigMapID}
type GetRequest struct{}

// GetResponse represents the response type
type GetResponse struct {
	*DeviceConfigMap
}

// ListResponse represents the response type
// Endpoint: GET /DeviceConfigMaps
type ListResponse struct {
	DeviceConfigMaps []*DeviceConfigMap `json:"deviceConfigMaps"`
}

// DeleteResponse represents the response type
type DeleteResponse struct{}
