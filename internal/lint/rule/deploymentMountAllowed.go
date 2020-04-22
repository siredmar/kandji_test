package rule

import (
	"fmt"

	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/result"
	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/errors"
)

// DeploymentMountAllowed passes iff the specified mounts are allowed.
type DeploymentMountAllowed struct {
	exclude map[string]bool
}

// NewDeploymentMountAllowed returns a new DeploymentMountAllowed rule,
// checking specified mounts against exclusions
func NewDeploymentMountAllowed(exclude []string) *DeploymentMountAllowed {
	m := make(map[string]bool)
	for _, e := range exclude {
		m[e] = true
	}
	return &DeploymentMountAllowed{
		exclude: m,
	}
}

// ID returns the ID of this rule
func (r *DeploymentMountAllowed) ID() string {
	return "DeploymentMountAllowed"
}

// Desc returns the description of this rule
func (r *DeploymentMountAllowed) Desc() string {
	var exclude []string
	for ex := range r.exclude {
		exclude = append(exclude, ex)
	}
	return fmt.Sprintf("The mounts of a deployment must not include %v", exclude)
}

// Exec checks compliance of the given resource with the rule
func (r *DeploymentMountAllowed) Exec(ctx *context.Context, resource interface{}) (*result.Result, error) {
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

	containers := res.Spec.Template.Spec.Containers

	if containers == nil || len(containers) == 0 {
		result.Skip = true
		result.Have = "no containers"
	}

	sawMount := false
	for _, c := range containers {
		for _, vol := range c.VolumeMounts {
			sawMount = true
			if _, exists := r.exclude[vol.MountPath]; exists {
				result.Pass = false
				result.Have = vol.MountPath
				return result, nil
			}
		}

	}

	if !sawMount {
		result.Skip = true
		result.Have = "no volumeMounts"
	}

	return result, nil
}
