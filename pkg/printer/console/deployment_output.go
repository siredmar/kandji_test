package printer

import (
	api "github.com/grid-x/gxctl/pkg/api"
)

type DeploymentConsoleOutput struct {
	ID            string `header:"ID"`
	App           string `header:"Application"`
	LastUpdatedAt string `header:"Last updated at"`
}

func (o DeploymentConsoleOutput) Map(d api.Deployment) DeploymentConsoleOutput {
	o.ID = d.Metadata.ID
	o.App = d.Spec.App
	o.LastUpdatedAt = d.Status.LastUpdatedAt

	return o
}
