package state

import (
	"github.com/grid-x/gxctl/pkg/api"
)

// State contains all resources that are relevant to linting
type State struct {
	Applications []api.Application
	Deployments  []api.Deployment
	Devices      []api.Device
}
