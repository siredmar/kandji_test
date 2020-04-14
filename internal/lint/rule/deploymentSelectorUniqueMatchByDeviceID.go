package rule

import (
	"github.com/grid-x/gxctl/internal/lint/result"
	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
)

// DeploymentSelectorUniqueMatchByDeviceID passes iff there is
// no existing deployment with the same matchByDeviceID selector
type DeploymentSelectorUniqueMatchByDeviceID struct {
	client *client.APIClient
}

// NewDeploymentSelectorUniqueMatchByDeviceID returns a new DeploymentSelectorUniqueMatchByDeviceID rule
func NewDeploymentSelectorUniqueMatchByDeviceID(client *client.APIClient) *DeploymentSelectorUniqueMatchByDeviceID {
	return &DeploymentSelectorUniqueMatchByDeviceID{
		client: client,
	}
}

// ID returns the ID of this rule
func (r *DeploymentSelectorUniqueMatchByDeviceID) ID() string {
	return "DeploymentSelectorUniqueMatchByDeviceID"
}

// Desc returns the description of this rule
func (r *DeploymentSelectorUniqueMatchByDeviceID) Desc() string {
	return "There must not be another deployment with the same matchByDeviceID selector"
}

// Exec checks compliance of the given resource with the rule
func (r *DeploymentSelectorUniqueMatchByDeviceID) Exec(resource interface{}) (*result.Result, error) {
	res, ok := resource.(api.Deployment)
	if !ok {
		return nil, errors.E(
			errors.Internal,
			"Wrong resource type",
		)
	}

	deployments, err := getDeployments(r.client)
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			"Can't get deployments",
		)
	}

	result := &result.Result{
		Pass: true,
	}

	ID := res.Spec.Selector.MatchByDeviceID
	if ID == nil {
		return result, nil
	}

	for _, d := range deployments.Deployments {
		if dID := d.Spec.Selector.MatchByDeviceID; dID != nil && dID == ID {
			result.Pass = false
			result.Have = dID
			break
		}
	}

	return result, nil
}

func getDeployments(client *client.APIClient) (api.Deployments, error) {
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
