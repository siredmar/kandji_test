package filter

import "github.com/grid-x/gxctl/pkg/api"

type Filter interface {
	// Eval should return true iff the input should be included.
	Eval(interface{}) (bool, error)
}

func Devices(devices api.Devices, f Filter) api.Devices {
	var out api.Devices

	for _, d := range devices.Devices {
		include, err := f.Eval(d)
		if err != nil || !include {
			continue
		}
		out.Devices = append(out.Devices, d)
	}

	return out
}
