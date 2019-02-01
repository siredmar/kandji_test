package v20181127

import (
	"github.com/grid-x/ds-api/api/management/2018-11-27/application"
)

// Service implements the service for API version 2018-11-27
type Service struct {
	Application *application.Service
}

// NewService creates a new service
func NewService(injections ...interface{}) (*Service, error) {
	return &Service{
		Application: application.NewService(injections...),
	}, nil
}
