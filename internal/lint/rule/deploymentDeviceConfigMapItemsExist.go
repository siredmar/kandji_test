package rule

import (
	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/result"
	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/errors"
)

// DeploymentDeviceConfigMapItemsExist passes iff all deviceConfigMap items referenced by volumeSources exist.
type DeploymentDeviceConfigMapItemsExist struct{}

// ID returns the ID of this rule
func (r *DeploymentDeviceConfigMapItemsExist) ID() string {
	return "DeploymentDeviceConfigMapItemsExist"
}

// Desc returns the description of this rule
func (r *DeploymentDeviceConfigMapItemsExist) Desc() string {
	return "All deviceConfigMaps items referenced by volumeSources must exist"
}

// Exec checks compliance of the given resource with the rule
func (r *DeploymentDeviceConfigMapItemsExist) Exec(ctx *context.Context, resource interface{}) (*result.Result, error) {
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

	want := make(map[string][]string)
	for _, v := range res.Spec.Template.Spec.Volumes {
		if v.ConfigMap == nil {
			continue
		}
		s := make([]string, 0)
		for _, k := range v.ConfigMap.Items {
			s = append(s, k.Key)
		}
		want[v.ConfigMap.Name] = s
	}

	have := make(map[string]map[string]struct{})
	for _, dcm := range ctx.DeviceConfigMaps() {
		m := make(map[string]struct{})
		for k := range dcm.Spec.Data {
			m[k] = struct{}{}
		}
		for k := range dcm.Spec.BinaryData {
			m[k] = struct{}{}
		}
		have[dcm.Metadata.ID] = m
	}

	for dcm, keys := range want {
		for _, k := range keys {
			if _, ok := have[dcm][k]; !ok {
				result.Pass = false
				result.Want = dcm + "." + k
				return result, nil
			}
		}
	}

	return result, nil
}
