package v20181102

import (
	"github.com/grid-x/ds-api-types"
)

// CreateRequest represents the request type
type CreateRequest struct {
	Spec DeviceSpec `json:"spec"`
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
	Metadata types.UpdateMetadata `json:"metadata"`
	Spec     UpdateSpec           `json:"spec,omitempty"`
}

// UpdateResponse represents the response type
type UpdateResponse struct {
	*Device
}

// DeleteResponse represents the response type
type DeleteResponse struct{}
