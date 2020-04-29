package rule

import (
	"fmt"

	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/result"
	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/errors"
)

// DeploymentVolumesAllowed passes iff the specified volumes (host path) are allowed.
type DeploymentVolumesAllowed struct {
	exclude map[string]bool
}

// NewDeploymentVolumesAllowed returns a new DeploymentVolumesAllowed rule,
// checking specified volumes against exclusions
func NewDeploymentVolumesAllowed(exclude []string) *DeploymentVolumesAllowed {
	m := make(map[string]bool)
	for _, e := range exclude {
		m[e] = true
	}
	return &DeploymentVolumesAllowed{
		exclude: m,
	}
}

// ID returns the ID of this rule
func (r *DeploymentVolumesAllowed) ID() string {
	return "DeploymentVolumesAllowed"
}

// Desc returns the description of this rule
func (r *DeploymentVolumesAllowed) Desc() string {
	var exclude []string
	for ex := range r.exclude {
		exclude = append(exclude, ex)
	}
	return fmt.Sprintf("The volumes of a deployment must not include %v", exclude)
}

// Exec checks compliance of the given resource with the rule
func (r *DeploymentVolumesAllowed) Exec(ctx *context.Context, resource interface{}) (*result.Result, error) {
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

	if volumes == nil || len(volumes) == 0 {
		result.Skip = true
		result.Have = "no volumes"
		return result, nil
	}

	for _, v := range volumes {
		if _, exists := r.exclude[v.VolumeSource.HostPath.Path]; exists {
			result.Pass = false
			result.Have = v.VolumeSource.HostPath.Path
			return result, nil
		}
	}

	return result, nil
}
