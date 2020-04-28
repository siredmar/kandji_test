package context

import (
	"fmt"
	"strings"

	"github.com/grid-x/gxctl/internal/lint/state"
	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
)

// Context contains both pending and current state
type Context struct {
	Current state.State
	Desired state.State
}

// New creates a new Context
func New(client *client.APIClient) (*Context, error) {
	a, err := fetchApplications(client)
	if err != nil {
		return nil, err
	}

	deps, err := fetchDeployments(client)
	if err != nil {
		return nil, err
	}

	devs, err := fetchDevices(client)
	if err != nil {
		return nil, err
	}

	current := state.State{
		Applications: a.Applications,
		Deployments:  deps.Deployments,
		Devices:      devs.Devices,
	}

	return &Context{
		Current: current,
	}, nil
}

// SetDesired sets the desired resources
func (c *Context) SetDesired(resources []interface{}) {
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

	c.Desired = state.State{
		Applications: apps,
		Deployments:  deps,
	}
}

func (c *Context) String() string {
	var str strings.Builder

	str.WriteString("Devices:\n")
	str.WriteString("  Current:\n")
	for _, x := range c.Current.Devices {
		str.WriteString(fmt.Sprintf("    %v\n", x.Metadata.ID))
	}
	str.WriteString("Applications:\n")
	str.WriteString("  Current:\n")
	for _, x := range c.Current.Applications {
		str.WriteString(fmt.Sprintf("    %v\n", x.Name))
	}
	str.WriteString("  Desired:\n")
	for _, x := range c.Desired.Applications {
		str.WriteString(fmt.Sprintf("    %v\n", x.Name))
	}
	str.WriteString("Deployments:\n")
	str.WriteString("  Current:\n")
	for _, x := range c.Current.Deployments {
		str.WriteString(fmt.Sprintf("    %v\n", x.Metadata.ID))
	}
	str.WriteString("  Desired:\n")
	for _, x := range c.Desired.Deployments {
		str.WriteString(fmt.Sprintf("    %v\n", x.Metadata.ID))
	}

	return str.String()
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

// Devices returns all devices
func (c *Context) Devices() []api.Device {
	return c.Current.Devices
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

	return deploymentList, nil
}

func fetchDevices(client *client.APIClient) (api.Devices, error) {
	response, err := client.GetRequest(api.DevicesEndpoint)
	if err != nil {
		return api.Devices{}, err
	}

	deviceList, err := api.NewDevices(response)
	if err != nil {
		return deviceList, err
	}

	return deviceList, nil
}
