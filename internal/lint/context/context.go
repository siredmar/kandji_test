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
	Cl      Client
}

// Client encapsulates the API methods (besides current and desired state) that linting rules need.
// It is normally implemented in terms of *client.APIClient, but wrapped for mocking in testing.
type Client interface {
	GetDeviceById(id string) (api.Device, error)
}

type clientImpl struct {
	*client.APIClient
}

func (c clientImpl) GetDeviceById(id string) (api.Device, error) {
	return client.GetDeviceById(c.APIClient, id, nil)
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

	dcms, err := fetchDeviceConfigMaps(client)
	if err != nil {
		return nil, err
	}

	current := state.State{
		Applications:     a.Applications,
		Deployments:      deps.Deployments,
		DeviceConfigMaps: dcms.DeviceConfigMaps,
	}

	return &Context{
		Current: current,
		Cl:      clientImpl{client},
	}, nil
}

// SetDesired sets the desired resources
func (c *Context) SetDesired(resources []interface{}) {
	var apps []api.Application
	var deps []api.Deployment
	var dcms []api.DeviceConfigMap

	for _, res := range resources {
		switch v := res.(type) {
		case *api.Application:
			apps = append(apps, *v)
		case *api.Deployment:
			deps = append(deps, *v)
		case *api.DeviceConfigMap:
			dcms = append(dcms, *v)
		}
	}

	c.Desired = state.State{
		Applications:     apps,
		Deployments:      deps,
		DeviceConfigMaps: dcms,
	}
}

func (c *Context) String() string {
	var str strings.Builder

	str.WriteString("DeviceConfigMaps:\n")
	str.WriteString("  Current:\n")
	for _, x := range c.Current.DeviceConfigMaps {
		str.WriteString(fmt.Sprintf("    %v\n", x.Metadata.ID))
	}
	str.WriteString("  Desired:\n")
	for _, x := range c.Desired.DeviceConfigMaps {
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

// DeviceConfigMaps returns all deployments
func (c *Context) DeviceConfigMaps() []api.DeviceConfigMap {
	var dcms []api.DeviceConfigMap
	dcms = append(dcms, c.Current.DeviceConfigMaps...)
	dcms = append(dcms, c.Desired.DeviceConfigMaps...)
	return dcms
}

func fetchApplications(client *client.APIClient) (api.Applications, error) {
	response, err := client.GetRequest(api.ApplicationsEndpoint)
	if err != nil {
		return api.Applications{}, err
	}

	applicationList, err := api.NewApplications(response, false)
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

	deploymentList, err := api.NewDeployments(response, false)
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

	deviceList, err := api.NewDevices(response, false)
	if err != nil {
		return deviceList, err
	}

	return deviceList, nil
}

func fetchDeviceConfigMaps(client *client.APIClient) (api.DeviceConfigMaps, error) {
	response, err := client.GetRequest(api.DeviceConfigMapsEndpoint)
	if err != nil {
		return api.DeviceConfigMaps{}, err
	}

	dcms, err := api.NewDeviceConfigMaps(response, false)
	if err != nil {
		return api.DeviceConfigMaps{}, err
	}

	return dcms, nil
}
