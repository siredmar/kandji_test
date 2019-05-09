package v20181128

import (
	"github.com/grid-x/ds-api-types"
)

// CreateRequest represents the request type
type CreateRequest struct {
	Spec *DeviceDeploymentSpec `json:"spec"`
}

// CreateResponse represents the response type
type CreateResponse struct {
	*Deployment
}

// ListResponse represents the response type
type ListResponse struct {
	Deployments []*Deployment `json:"deployments"`
}

// GetResponse represents the response type
type GetResponse struct {
	*Deployment
}

// UpdateRequest represents the request type
type UpdateRequest struct {
	Metadata types.UpdateMetadata  `json:"metadata"`
	Spec     *DeviceDeploymentSpec `json:"spec"`
}

// UpdateResponse represents the response type
type UpdateResponse struct {
	*Deployment
}

// DeleteResponse represents the response type
type DeleteResponse struct{}
