package rule

import (
	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/result"
	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/errors"
)

// DeploymentVolumesExist passes iff all volumes referenced by volumeMounts exist.
type DeploymentVolumesExist struct{}

// ID returns the ID of this rule
func (r *DeploymentVolumesExist) ID() string {
	return "DeploymentVolumesExist"
}

// Desc returns the description of this rule
func (r *DeploymentVolumesExist) Desc() string {
	return "All volumes referenced by volumeMounts must exist"
}

// Exec checks compliance of the given resource with the rule
func (r *DeploymentVolumesExist) Exec(ctx *context.Context, resource interface{}) (*result.Result, error) {
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
	if containers == nil || len(containers) == 0 {
		result.Skip = true
		result.Have = "no containers"
	}

	volumes := make(map[string]bool)
	for _, v := range res.Spec.Template.Spec.Volumes {
		volumes[v.Name] = true
	}

	haveVolumeMounts := false
	for _, c := range containers {
		for _, v := range c.VolumeMounts {
			haveVolumeMounts = true
			if _, exists := volumes[v.Name]; !exists {
				result.Pass = false
				result.Have = v.Name
				return result, nil
			}
		}
	}

	if !haveVolumeMounts {
		result.Skip = true
		result.Have = "no volumeMounts"
	}

	return result, nil
}
