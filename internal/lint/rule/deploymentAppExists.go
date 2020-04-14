package rule

import (
	"github.com/grid-x/gxctl/internal/lint/result"
	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
)

// DeploymentAppExists passes iff the specified app exists.
type DeploymentAppExists struct {
	client *client.APIClient
}

// NewDeploymentAppExists returns a new DeploymentAppExists rule
func NewDeploymentAppExists(client *client.APIClient) *DeploymentAppExists {
	return &DeploymentAppExists{
		client: client,
	}
}

// ID returns the ID of this rule
func (r *DeploymentAppExists) ID() string {
	return "DeploymentAppExists"
}

// Desc returns the description of this rule
func (r *DeploymentAppExists) Desc() string {
	return "The app of a deployment must exist"
}

// Exec checks compliance of the given resource with the rule
func (r *DeploymentAppExists) Exec(resource interface{}) (*result.Result, error) {
	res, ok := resource.(api.Deployment)
	if !ok {
		return nil, errors.E(
			errors.Internal,
			"Wrong resource type",
		)
	}

	app := res.Spec.App
	result := &result.Result{
		Have: app,
	}
	apps, err := getApplications(r.client)
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			"Can't get applications",
		)
	}
	var names []string
	for _, a := range apps.Applications {
		names = append(names, a.Name)
		if a.Name == app {
			result.Pass = true
			break
		}
	}
	result.Want = names

	return result, nil
}

func getApplications(client *client.APIClient) (api.Applications, error) {
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
