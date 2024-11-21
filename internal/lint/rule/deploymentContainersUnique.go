package rule

import (
	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/result"
	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/errors"
)

// DeploymentContainersUnique passes iff the specified containers are unique
type DeploymentContainersUnique struct{}

// NewDeploymentContainersUnique returns a new DeploymentContainersUnique rule
func NewDeploymentContainersUnique() *DeploymentContainersUnique {
	return &DeploymentContainersUnique{}
}

// ID returns the ID of this rule
func (r *DeploymentContainersUnique) ID() string {
	return "DeploymentContainersUnique"
}

// Desc returns the description of this rule
func (r *DeploymentContainersUnique) Desc() string {
	return "The containers of a deployment must be unique"
}

// Exec checks compliance of the given resource with the rule
func (r *DeploymentContainersUnique) Exec(ctx *context.Context, resource interface{}) (*result.Result, error) {

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

	cn := make(map[string]bool)

	if len(containers) == 0 {
		result.Skip = true
		result.Have = "no containers"
		return result, nil
	}

	for _, c := range containers {
		if _, exists := cn[c.Name]; exists {
			result.Pass = false
			result.Have = c.Name
			return result, nil
		}
		cn[c.Name] = true
	}

	return result, nil
}
