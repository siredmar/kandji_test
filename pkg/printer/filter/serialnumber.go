package filter

import (
	"strings"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/errors"
)

type SerialnumberFilter struct {
	serialnumber string
}

func NewSerialnumberFilter(serialnumber string) *SerialnumberFilter {
	return &SerialnumberFilter{
		serialnumber:     serialnumber,
	}
}

func (f *SerialnumberFilter) Eval(in interface{}) (bool, error) {
	device, ok := in.(api.Device)

	if !ok {
		return false, errors.E(
			errors.Internal,
			"Wrong resource type",
		)
	}

	if f.serialnumber == "" {
		return true, nil
	}

	return strings.Contains(device.Spec.Serialnumber, f.serialnumber), nil
}
