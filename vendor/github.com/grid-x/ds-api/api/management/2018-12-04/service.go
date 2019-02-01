package v20181204

import (
	"github.com/grid-x/ds-api/api/management/2018-12-04/account"
	"github.com/grid-x/ds-api/api/management/2018-12-04/maintenance"
	"github.com/grid-x/ds-api/api/management/2018-12-04/pods"
	"github.com/grid-x/ds-api/api/management/2018-12-04/users"
)

// Service implements the service for API version 2018-12-04
type Service struct {
	Maintenance *maintenance.Service
	Pods        *pods.Service
	Account     *account.Service
	Users       *users.Service
}

// NewService creates a new service
func NewService(injections ...interface{}) (*Service, error) {
	return &Service{
		Maintenance: maintenance.NewService(injections...),
		Pods:        pods.NewService(injections...),
		Account:     account.NewService(injections...),
		Users:       users.NewService(injections...),
	}, nil
}
