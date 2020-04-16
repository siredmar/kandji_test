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

	ids := make(map[string]bool)
	switch res := resource.(type) {

	case api.Application:
		if res.Name == "" {
			result.Skip = true
			result.Have = "Name not set"
			return result, nil
		}
		applications := ctx.Applications()
		for _, a := range applications {
			ID := a.Name
			if _, exists := ids[ID]; ID != "" && exists {
				result.Pass = false
				result.Have = ID
				return result, nil
			}
			if ID != "" {
				ids[ID] = true
			}
		}

	case api.Deployment:
		if res.Metadata.ID == "" {
			result.Skip = true
			result.Have = "Metadata.ID not set"
			return result, nil
		}
		deployments := ctx.Deployments()
		for _, d := range deployments {
			ID := d.Metadata.ID
			if _, exists := ids[ID]; ID != "" && exists {
				result.Pass = false
				result.Have = ID
				return result, nil
			}
			if ID != "" {
				ids[ID] = true
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
