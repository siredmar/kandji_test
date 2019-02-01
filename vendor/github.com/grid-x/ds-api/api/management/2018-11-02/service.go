package v20181102

import (
	"github.com/grid-x/ds-api/api/management/2018-11-02/devices"
)

// Service implements the service for API version 2018-11-02
type Service struct {
	Devices *devices.Service
}

// NewService creates a new service
func NewService(injections ...interface{}) (*Service, error) {
	return &Service{
		Devices: devices.NewService(injections...),
	}, nil
}
