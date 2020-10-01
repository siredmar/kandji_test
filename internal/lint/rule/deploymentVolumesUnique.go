package rule

import (
	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/result"
	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/errors"
)

// DeploymentVolumesUnique passes iff the specified volume mounts are unique.
type DeploymentVolumesUnique struct {}

// NewDeploymentVolumesUnique returns a new DeploymentVolumesUnique rule
func NewDeploymentVolumesUnique() *DeploymentVolumesUnique {
	return &DeploymentVolumesUnique{}
}

// ID returns the ID of this rule
func (r *DeploymentVolumesUnique) ID() string {
	return "DeploymentVolumesUnique"
}

// Desc returns the description of this rule
func (r *DeploymentVolumesUnique) Desc() string {
	return "The volumes of a deployment must be unique"
}

// Exec checks compliance of the given resource with the rule
func (r *DeploymentVolumesUnique) Exec(ctx *context.Context, resource interface{}) (*result.Result, error) {

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

	volumes := res.Spec.Template.Spec.Volumes

	vn := make(map[string]bool)

	if volumes == nil || len(volumes) == 0 {
		result.Skip = true
		result.Have = "no volumes"
		return result, nil
	}

	for _, v := range volumes {
		if _, exists := vn[v.Name]; exists {
			result.Pass = false
			result.Have = v.Name
			return result, nil
		}
		vn[v.Name] = true
	}

	return result, nil
}
