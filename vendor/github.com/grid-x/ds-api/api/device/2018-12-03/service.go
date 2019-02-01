package v20181203

import (
	"github.com/grid-x/ds-api/api/device/2018-12-03/device"
	"github.com/grid-x/ds-api/api/device/2018-12-03/pods"
)

// Service implements the service for API version 2018-12-03
type Service struct {
	Device *device.Service
	Pods   *pods.Service
}

// NewService creates a new service
func NewService(injections ...interface{}) (*Service, error) {
	return &Service{
		Device: device.NewService(injections...),
		Pods:   pods.NewService(injections...),
	}, nil
}
