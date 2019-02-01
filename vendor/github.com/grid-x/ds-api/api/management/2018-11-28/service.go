package v20181128

import (
	"github.com/grid-x/ds-api/api/management/2018-11-28/deployments"
)

// Service implements the service for API version 2018-11-28
type Service struct {
	Deployments *deployments.Service
}

// NewService creates a new service
func NewService(injections ...interface{}) (*Service, error) {
	return &Service{
		Deployments: deployments.NewService(injections...),
	}, nil
}
