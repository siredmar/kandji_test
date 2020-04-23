package rule

import (
	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/result"
	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/errors"
)

// DeploymentAppExists passes iff the specified app exists.
type DeploymentAppExists struct{}

// ID returns the ID of this rule
func (r *DeploymentAppExists) ID() string {
	return "DeploymentAppExists"
}

// Desc returns the description of this rule
func (r *DeploymentAppExists) Desc() string {
	return "The app of a deployment must exist"
}

// Exec checks compliance of the given resource with the rule
func (r *DeploymentAppExists) Exec(ctx *context.Context, resource interface{}) (*result.Result, error) {
	res, ok := resource.(api.Deployment)
	if !ok {
		return nil, errors.E(
			errors.Internal,
			"Wrong resource type",
		)
	}

	app := res.Spec.App
	apps := ctx.Applications()

	if len(apps) == 0 {
		return &result.Result{
			Pass: true,
			Skip: true,
			Have: "no applications",
		}, nil
	}

	result := &result.Result{}
	var names []string
	for _, a := range apps {
		names = append(names, a.Name)
		if a.Name == app {
			result.Pass = true
			result.Have = app
			break
		}
	}
	result.Want = names

	return result, nil
}
