package rule

import (
	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/result"
	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/errors"
)

// DeploymentSelectorDeviceExists passes iff the device specified
// by MatchByDeviceID currently exists
type DeploymentSelectorDeviceExists struct{}

// ID returns the ID of this rule
func (r *DeploymentSelectorDeviceExists) ID() string {
	return "DeploymentSelectorDeviceExists"
}

// Desc returns the description of this rule
func (r *DeploymentSelectorDeviceExists) Desc() string {
	return "A deployment with MatchByDeviceID selector must specify an existing device"
}

// Exec checks compliance of the given resource with the rule
func (r *DeploymentSelectorDeviceExists) Exec(ctx *context.Context, resource interface{}) (*result.Result, error) {
	res, ok := resource.(*api.Deployment)
	if !ok {
		return nil, errors.E(
			errors.Internal,
			"Wrong resource type",
		)
	}

	result := &result.Result{
		Pass: true,
	}

	selector := res.Spec.Selector.MatchByDeviceID
	if selector == nil {
		result.Skip = true
		result.Have = "no MatchByDeviceID"
		return result, nil
	}

	_, err := ctx.Cl.GetDeviceByID(*selector)
	if err != nil {
		result.Pass = false
		result.Have = *selector
		return result, nil
	}

	return result, nil
}

func (r *DeploymentSelectorDeviceExists) GetDeviceByID() error {
	return nil
}
