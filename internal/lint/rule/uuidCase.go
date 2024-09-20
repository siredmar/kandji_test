package rule

import (
	"fmt"
	"regexp"

	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/result"
	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/errors"
)

const (
	uuidRegex = "^[0-9a-f]{8}(-[0-9a-f]{4}){3}-[0-9a-f]{12}$"
)

// UUIDCase passes iff uuids are all lowercase
type UUIDCase struct {
	path string
}

// ID returns the ID of this rule
func (r *UUIDCase) ID() string {
	return "UUIDCase"
}

// Desc returns the description of this rule
func (r *UUIDCase) Desc() string {
	return fmt.Sprintf("%s must be a valid lowercase UUID. Regex used for validation is '%s'", r.path, uuidRegex)
}

func (r *UUIDCase) checkRegex(uuid string) (*result.Result, error) {
	result := &result.Result{
		Pass: true,
		Skip: true,
	}
	match, err := regexp.MatchString(uuidRegex, uuid)
	if err != nil {
		return nil, errors.E(
			errors.Internal,
			"Error while matching UUID",
		)
	}
	if !match {
		result.Skip = false
		result.Pass = false
		result.Have = uuid
		return result, nil
	}
	return result, nil
}

// Exec checks compliance of the given resource with the rule
func (r *UUIDCase) Exec(ctx *context.Context, resource interface{}) (*result.Result, error) {
	var result *result.Result
	var err error
	r.path = "metadata.ID"
	result, err = r.checkRegex((resource.(api.Resource)).Meta().ID)
	if err != nil {
		return nil, err
	}
	if !result.Pass {
		return result, nil
	}

	switch res := resource.(type) {
	case *api.MaintenanceTask:
		r.path = "res.spec.deviceID"
		result, err = r.checkRegex(res.Spec.DeviceID)
	case *api.Deployment:
		if res.Spec.Selector.MatchByDeviceID != nil {
			deviceID := res.Spec.Selector.MatchByDeviceID
			r.path = "spec.selector.matchByDeviceID"
			result, err = r.checkRegex(*deviceID)
		}

	default:
		return nil, errors.E(
			errors.Internal,
			"Wrong resource type",
		)
	}

	return result, err
}
