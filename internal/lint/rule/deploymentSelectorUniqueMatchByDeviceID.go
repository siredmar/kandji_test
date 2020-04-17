package rule

import (
	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/result"
	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/errors"
)

// DeploymentSelectorUniqueMatchByDeviceID passes iff there is
// no existing deployment with the same matchByDeviceID selector
type DeploymentSelectorUniqueMatchByDeviceID struct{}

// ID returns the ID of this rule
func (r *DeploymentSelectorUniqueMatchByDeviceID) ID() string {
	return "DeploymentSelectorUniqueMatchByDeviceID"
}

// Desc returns the description of this rule
func (r *DeploymentSelectorUniqueMatchByDeviceID) Desc() string {
	return "There must not be another deployment with the same matchByDeviceID selector"
}

// Exec checks compliance of the given resource with the rule
func (r *DeploymentSelectorUniqueMatchByDeviceID) Exec(ctx *context.Context, resource interface{}) (*result.Result, error) {
	res, ok := resource.(api.Deployment)
	if !ok {
		return nil, errors.E(
			errors.Internal,
			"Wrong resource type",
		)
	}

	deployments := ctx.Deployments()

	result := &result.Result{
		Pass: true,
	}

	ID := res.Spec.Selector.MatchByDeviceID
	if ID == nil {
		return result, nil
	}

	for _, d := range deployments {
		if dID := d.Spec.Selector.MatchByDeviceID; dID != nil && *dID == *ID {
			result.Pass = false
			result.Have = *dID
			break
		}
	}

	return result, nil
}
