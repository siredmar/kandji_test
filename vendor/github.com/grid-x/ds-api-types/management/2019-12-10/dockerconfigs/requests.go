package v20191210

import (
	types "github.com/grid-x/ds-api-types"
)

// CreateRequest represents the request type
// Endpoint: POST /dockerconfigs
type CreateRequest struct {
	Metadata types.Metadata    `json:"metadata"`
	Spec     *DockerConfigSpec `json:"spec"`
}

// CreateResponse represents the response type
type CreateResponse struct {
	*DockerConfig
}

// UpdateRequest represents the request type
// Endpoint: PATCH /dockerconfigs/{dockerConfigID}
type UpdateRequest struct {
	Metadata types.UpdateMetadata `json:"metadata"`
	Spec     *DockerConfigSpec    `json:"spec"`
}

// UpdateResponse represents the response type
type UpdateResponse struct {
	*DockerConfig
}

// GetRequest represents the request type
// Endpoint: GET /dockerconfigs/{dockerConfigID}
type GetRequest struct{}

// GetResponse represents the response type
// Endpoint: GET /dockerconfigs/{dockerConfigID}
type GetResponse struct {
	*DockerConfig
}

// ListResponse represents the response type
// Endpoint: GET /dockerconfigs
type ListResponse struct {
	DockerConfigs []*DockerConfig `json:"dockerConfigs"`
}

// DeleteResponse represents the response type
type DeleteResponse struct{}
