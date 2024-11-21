package rule

import (
	types "github.com/grid-x/ds-api-types"
	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/result"
	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/errors"
)

// DeploymentImageRefValid passes iff the specified containers are unique
type DeploymentImageRefValid struct{}

// NewDeploymentImageRefValid returns a new DeploymentImageRefValid rule
func NewDeploymentImageRefValid() *DeploymentImageRefValid {
	return &DeploymentImageRefValid{}
}

// ID returns the ID of this rule
func (r *DeploymentImageRefValid) ID() string {
	return "DeploymentImageRefValid"
}

// Desc returns the description of this rule
func (r *DeploymentImageRefValid) Desc() string {
	return "The image references of a deployment must be valid"
}

// Exec checks compliance of the given resource with the rule
func (r *DeploymentImageRefValid) Exec(ctx *context.Context, resource interface{}) (*result.Result, error) {

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

	containers := res.Spec.Template.Spec.Containers

	if len(containers) == 0 {
		result.Skip = true
		result.Have = "no containers"
		return result, nil
	}

	for _, c := range containers {
		if !types.IsValidImageRef(c.Image) {
			result.Pass = false
			result.Have = c.Image
			return result, nil
		}
	}

	return result, nil
}
