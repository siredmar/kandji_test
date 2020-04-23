package rule

import (
	"fmt"

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

	result := &result.Result{
		Pass: true,
	}

	deployments := ctx.Deployments()
	if len(deployments) == 0 {
		result.Skip = true
		result.Have = "no deployments"
		return result, nil
	}

	ID := res.Spec.Selector.MatchByDeviceID
	app := res.Spec.App
	if ID == nil {
		result.Skip = true
		result.Have = "no matchByDeviceID"
		return result, nil
	}

	for _, d := range deployments {
		if dID := d.Spec.Selector.MatchByDeviceID; res.Metadata.ID != d.Metadata.ID && dID != nil && *dID == *ID {
			if dApp := d.Spec.App; dApp == app {
				result.Pass = false
				result.Have = fmt.Sprintf("deployment %s with app %s", *dID, dApp)
				break
			}
		}
	}

	return result, nil
}
