package context

import (
	"github.com/grid-x/gxctl/internal/lint/state"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
)

// Context contains both pending and current state
type Context struct {
	Current state.State
	Desired state.State
}

// New creates a new Context
func New(client *client.APIClient, resources []interface{}) (*Context, error) {
	a, err := fetchApplications(client)
	if err != nil {
		return nil, err
	}

	d, err := fetchDeployments(client)
	if err != nil {
		return nil, err
	}

	current := state.State{
		Applications: a.Applications,
		Deployments:  d.Deployments,
	}

	var apps []api.Application
	var deps []api.Deployment

	for _, res := range resources {
		switch v := res.(type) {
		case api.Application:
			apps = append(apps, v)
		case api.Deployment:
			deps = append(deps, v)
		}
	}

	desired := state.State{
		Applications: apps,
		Deployments:  deps,
	}

	return &Context{
		Current: current,
		Desired: desired,
	}, nil
}

// Applications returns all applications
func (c *Context) Applications() []api.Application {
	var apps []api.Application
	apps = append(apps, c.Current.Applications...)
	apps = append(apps, c.Desired.Applications...)
	return apps
}

// Deployments returns all deployments
func (c *Context) Deployments() []api.Deployment {
	var deps []api.Deployment
	deps = append(deps, c.Current.Deployments...)
	deps = append(deps, c.Desired.Deployments...)
	return deps
}

func fetchApplications(client *client.APIClient) (api.Applications, error) {
	response, err := client.GetRequest(api.ApplicationsEndpoint)
	if err != nil {
		return api.Applications{}, err
	}

	applicationList, err := api.NewApplications(response)
	if err != nil {
		return applicationList, err
	}

	if applicationList.IsEmpty() {
		return applicationList, errors.E(
			errors.NotExists,
			"no applications found",
		)
	}

	return applicationList, nil
}

func fetchDeployments(client *client.APIClient) (api.Deployments, error) {
	response, err := client.GetRequest(api.DeploymentsEndpoint)
	if err != nil {
		return api.Deployments{}, err
	}

	deploymentList, err := api.NewDeployments(response)
	if err != nil {
		return deploymentList, err
	}

	if deploymentList.IsEmpty() {
		return deploymentList, errors.E(
			errors.NotExists,
			"no deployments found",
		)
	}

	return deploymentList, nil
}
