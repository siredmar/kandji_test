package rule

import (
	"github.com/google/go-cmp/cmp"

	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/result"
	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/errors"
)

// ResourceUniqueID passes iff there is
// no other resource of the same type with the same ID and a different spec
type ResourceUniqueID struct{}

// ID returns the ID of this rule
func (r *ResourceUniqueID) ID() string {
	return "ResourceUniqueID"
}

// Desc returns the description of this rule
func (r *ResourceUniqueID) Desc() string {
	return "There must not be another resource of the same type with the same ID and a different spec"
}

// Exec checks compliance of the given resource with the rule
func (r *ResourceUniqueID) Exec(ctx *context.Context, resource interface{}) (*result.Result, error) {
	result := &result.Result{
		Pass: true,
	}

	resources := make(map[string]interface{})

	switch res := resource.(type) {
	case api.Application:
		ID := res.Name
		if ID == "" {
			result.Skip = true
			result.Have = "Name not set"
			return result, nil
		}
		applications := append(ctx.Applications())
		for _, a := range applications {
			resources[a.Name] = res
		}
		if _, exists := resources[ID]; exists {
			result.Pass = false
			result.Have = ID
			return result, nil
		}
		break

	case api.Deployment:
		ID := res.Metadata.ID
		if ID == "" {
			result.Skip = true
			result.Have = "Metadata.ID not set"
			return result, nil
		}
		deployments := append(ctx.Deployments())
		for _, d := range deployments {
			resources[d.Metadata.ID] = d
		}
		_x, exists := resources[ID]
		if !exists {
			break
		}
		x := _x.(api.Deployment)
		if !cmp.Equal(res.Spec, x.Spec) {
			result.Pass = false
			result.Have = ID
			return result, nil
		}
		break

	default:
		return nil, errors.E(
			errors.Internal,
			"Wrong resource type",
		)
	}

	return result, nil
}
