package rule

import (
	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/result"
	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/errors"
)

// DeploymentDeviceConfigMapExist passes iff all deviceConfigMaps referenced by volumeSources exist.
type DeploymentDeviceConfigMapExist struct{}

// ID returns the ID of this rule
func (r *DeploymentDeviceConfigMapExist) ID() string {
	return "DeploymentDeviceConfigMapExist"
}

// Desc returns the description of this rule
func (r *DeploymentDeviceConfigMapExist) Desc() string {
	return "All deviceConfigMaps referenced by volumeSources must exist"
}

// Exec checks compliance of the given resource with the rule
func (r *DeploymentDeviceConfigMapExist) Exec(ctx *context.Context, resource interface{}) (*result.Result, error) {
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

	want := make(map[string]struct{})
	for _, v := range res.Spec.Template.Spec.Volumes {
		if v.ConfigMap == nil {
			continue
		}
		want[v.ConfigMap.Name] = struct{}{}
	}

	have := make(map[string]struct{})
	for _, dcm := range ctx.DeviceConfigMaps() {
		have[dcm.Metadata.ID] = struct{}{}
	}

	for dcm := range want {
		if _, ok := have[dcm]; !ok {
			result.Pass = false
			result.Want = dcm
			return result, nil
		}
	}

	return result, nil
}
