package v20181102

import (
	"github.com/grid-x/ds-api/api/device/2018-11-02/auth"
)

// Service implements the service for API version 2018-11-02
type Service struct {
	Auth *auth.Service
}

// NewService creates a new service
func NewService(injections ...interface{}) (*Service, error) {
	return &Service{
		Auth: auth.NewService(injections...),
	}, nil
}
