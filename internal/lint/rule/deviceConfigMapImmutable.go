package rule

import (
	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/result"
	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/errors"
)

// DeviceConfigMapImmutable passes iff all keys in the deviceConfigMap are unique
type DeviceConfigMapImmutable struct{}

// ID returns the ID of this rule
func (r *DeviceConfigMapImmutable) ID() string {
	return "DeviceConfigMapImmutable"
}

// Desc returns the description of this rule
func (r *DeviceConfigMapImmutable) Desc() string {
	return "An immutable deviceConfigMaps may not be modified"
}

// Exec checks compliance of the given resource with the rule
func (r *DeviceConfigMapImmutable) Exec(ctx *context.Context, resource interface{}) (*result.Result, error) {
	res, ok := resource.(*api.DeviceConfigMap)
	if !ok {
		return nil, errors.E(
			errors.Internal,
			"Wrong resource type",
		)
	}

	result := &result.Result{
		Pass: true,
	}

	have := ctx.DeviceConfigMaps()

	for _, dcm := range have {
		if dcm.Metadata.ID == res.Metadata.ID && dcm.Spec.Immutable != nil && *dcm.Spec.Immutable == true {
			result.Pass = false
			return result, nil
		}
	}

	return result, nil
}
