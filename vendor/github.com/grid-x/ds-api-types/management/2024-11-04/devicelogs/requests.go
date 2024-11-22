package v20241104

import types "github.com/grid-x/ds-api-types"

// GetRequest represents the request type
// Endpoint: GET /devicelogs/{deviceID}
type GetRequest struct{}

// GetResponse represents the response type
type GetResponse struct {
	*DeviceLogs
}

// DeleteRequest represents the request type
// Endpoint: DELETE /devicelogs/{deviceID}
type DeleteRequest struct{}

// DeleteResponse represents the response type
type DeleteResponse struct{}

// ListRequest represents the request type
// Endpoint: GET /devicelogs
type ListRequest struct{}

// ListResponse represents the response type
type ListResponse struct {
	*DeviceLogsList
}

// UpdateRequest represents the request type
// Endpoint: PATCH /devicelogs/{deviceID}
type UpdateRequest struct {
	Metadata types.UpdateMetadata `json:"metadata,omitempty"`
	Spec     UpdateSpec           `json:"spec,omitempty"`
}

// UpdateResponse represents the response type
type UpdateResponse struct {
	*DeviceLogs
}

// CreateRequest represents the request type
// Endpoint: POST /devicelogs/{deviceID}
type CreateRequest struct {
	Metadata types.Metadata `json:"metadata,omitempty"`
	Spec     CreateSpec     `json:"spec,omitempty"`
}

// CreateResponse represents the response type
type CreateResponse struct {
	*DeviceLogs
}
