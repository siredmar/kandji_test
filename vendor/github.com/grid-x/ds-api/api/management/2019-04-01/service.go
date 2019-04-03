package v20190401

import (
	"github.com/grid-x/ds-api/api/management/2019-04-01/dockerconfigs"
)

// Service implements the service for API version 2019-04-01
type Service struct {
	Dockerconfigs *dockerconfigs.Service
}

// NewService creates a new service
func NewService(injections ...interface{}) (*Service, error) {
	return &Service{
		Dockerconfigs: dockerconfigs.NewService(injections...),
	}, nil
}
