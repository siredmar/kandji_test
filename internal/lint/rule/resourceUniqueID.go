package rule

import (
	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/result"
	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/errors"
)

// ResourceUniqueID passes iff there is
// no other resource of the same type with the same ID
type ResourceUniqueID struct{}

// ID returns the ID of this rule
func (r *ResourceUniqueID) ID() string {
	return "ResourceUniqueID"
}

// Desc returns the description of this rule
func (r *ResourceUniqueID) Desc() string {
	return "There must not be another resource of the same type with the same ID"
}

// Exec checks compliance of the given resource with the rule
func (r *ResourceUniqueID) Exec(ctx *context.Context, resource interface{}) (*result.Result, error) {
	result := &result.Result{
		Pass: true,
	}

	switch res := resource.(type) {
	case *api.Application:
		ID := res.Name
		if ID == "" {
			result.Skip = true
			result.Have = "Name not set"
			return result, nil
		}
		applications := ctx.Desired.Applications
		if len(applications) == 0 {
			result.Skip = true
			result.Have = "no applications"
			return result, nil
		}
		for _, a := range applications {
			if a.Name == ID {
				result.Pass = false
				result.Have = ID
				break
			}
		}

	case *api.Deployment:
		ID := res.Metadata.ID
		if ID == "" {
			result.Skip = true
			result.Have = "Metadata.ID not set"
			return result, nil
		}
		deployments := ctx.Desired.Deployments
		if len(deployments) == 0 {
			result.Skip = true
			result.Have = "no deployments"
			return result, nil
		}
		for _, d := range deployments {
			if d.Metadata.ID == ID {
				result.Pass = false
				result.Have = ID
				break
			}
		}

	case *api.DeviceConfigMap:
		ID := res.Metadata.ID
		if ID == "" {
			result.Skip = true
			result.Have = "Metadata.ID not set"
			return result, nil
		}
		dcms := ctx.Desired.DeviceConfigMaps
		if len(dcms) == 0 {
			result.Skip = true
			result.Have = "no deviceConfigMaps"
			return result, nil
		}
		for _, d := range dcms {
			if d.Metadata.ID == ID {
				result.Pass = false
				result.Have = ID
				break
			}
		}

	default:
		return nil, errors.E(
			errors.Internal,
			"Wrong resource type",
		)
	}

	return result, nil
}
