package rule

import (
	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/result"
	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/errors"
)

// DeploymentSelectorSpecified passes iff there is
// a selector for this deployment.
type DeploymentSelectorSpecified struct{}

// ID returns the ID of this rule
func (r *DeploymentSelectorSpecified) ID() string {
	return "DeploymentSelectorSpecified"
}

// Desc returns the description of this rule
func (r *DeploymentSelectorSpecified) Desc() string {
	return "A deployment must have a MatchByDeviceID or MatchByLabel selector"
}

// Exec checks compliance of the given resource with the rule
func (r *DeploymentSelectorSpecified) Exec(ctx *context.Context, resource interface{}) (*result.Result, error) {
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
	if s := res.Spec.Selector; s.MatchByDeviceID == nil && len(s.MatchByLabels) == 0 {
		result.Pass = false
		result.Have = "no selector"
	}

	return result, nil
}
