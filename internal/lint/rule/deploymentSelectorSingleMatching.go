package rule

import (
	"github.com/grid-x/gxctl/internal/lint/result"
	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/errors"
)

// DeploymentSelectorSingleMatching passes iff there is
// at most one of MatchByDeviceID or MatchByLabel selector.
type DeploymentSelectorSingleMatching struct{}

// ID returns the ID of this rule
func (r *DeploymentSelectorSingleMatching) ID() string {
	return "DeploymentSelectorSingleMatching"
}

// Desc returns the decsription of this rule
func (r *DeploymentSelectorSingleMatching) Desc() string {
	return "A deployment must have at most one of MatchByDeviceID or MatchByLabel selectors"
}

// Exec checks compliance of the given resource with the rule
func (r *DeploymentSelectorSingleMatching) Exec(resource interface{}) (*result.Result, error) {
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
	if s := res.Spec.Selector; s.MatchByDeviceID != nil && len(s.MatchByLabels) > 0 {
		result.Pass = false
	}

	return result, nil
}
