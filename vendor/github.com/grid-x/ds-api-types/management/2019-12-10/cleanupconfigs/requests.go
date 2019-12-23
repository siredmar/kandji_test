package v20191210

import (
	types "github.com/grid-x/ds-api-types"
)

// CreateRequest represents the request type
// Endpoint: POST /cleanupconfigs
type CreateRequest struct {
	Metadata types.Metadata     `json:"metadata"`
	Spec     *CleanupConfigSpec `json:"spec"`
}

// CreateResponse represents the response type
type CreateResponse struct {
	*CleanupConfig
}

// UpdateRequest represents the request type
// Endpoint: PATCH /cleanupconfigs/{cleanupConfigID}
type UpdateRequest struct {
	Metadata types.UpdateMetadata `json:"metadata"`
	Spec     *CleanupConfigSpec   `json:"spec"`
}

// UpdateResponse represents the response type
type UpdateResponse struct {
	*CleanupConfig
}

// GetRequest represents the request type
// Endpoint: GET /cleanupconfigs/{cleanupConfigID}
type GetRequest struct{}

// GetResponse represents the response type
type GetResponse struct {
	*CleanupConfig
}

// ListResponse represents the response type
// Endpoint: GET /cleanupconfigs
type ListResponse struct {
	CleanupConfigs []*CleanupConfig `json:"cleanupConfigs"`
}

// DeleteResponse represents the response type
type DeleteResponse struct{}
