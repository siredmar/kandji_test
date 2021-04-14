package rule

import (
	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/result"
	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/errors"
)

// DeviceConfigMapUniqueKeys passes iff all keys in the deviceConfigMap are unique
type DeviceConfigMapUniqueKeys struct{}

// ID returns the ID of this rule
func (r *DeviceConfigMapUniqueKeys) ID() string {
	return "DeviceConfigMapUniqueKeys"
}

// Desc returns the description of this rule
func (r *DeviceConfigMapUniqueKeys) Desc() string {
	return "All deviceConfigMaps keys must be unique across data and binary data"
}

// Exec checks compliance of the given resource with the rule
func (r *DeviceConfigMapUniqueKeys) Exec(ctx *context.Context, resource interface{}) (*result.Result, error) {
	res, ok := resource.(api.DeviceConfigMap)
	if !ok {
		return nil, errors.E(
			errors.Internal,
			"Wrong resource type",
		)
	}

	result := &result.Result{
		Pass: true,
	}

	want := make(map[string]struct{})
	for k := range res.Spec.Data {
		want[k] = struct{}{}
	}

	for k := range res.Spec.BinaryData {
		if _, ok := want[k]; ok {
			result.Pass = false
			result.Have = k
			return result, nil
		}
		want[k] = struct{}{}
	}

	return result, nil
}
