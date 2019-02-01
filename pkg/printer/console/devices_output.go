package printer

import (
	"sort"

	api "github.com/grid-x/gxctl/pkg/api"
)

type DevicesConsoleOutput []DeviceConsoleOutput
type DevicesConsoleOutputWide []DeviceConsoleOutputWide

func (do DevicesConsoleOutput) Map(d api.Devices) DevicesConsoleOutput {
	var out DevicesConsoleOutput
	for _, e := range d.Devices {
		out = append(out, DeviceConsoleOutput{}.Map(e))
	}

	return out
}

func (do DevicesConsoleOutputWide) Map(d api.Devices) DevicesConsoleOutputWide {
	var out DevicesConsoleOutputWide
	for _, e := range d.Devices {
		out = append(out, DeviceConsoleOutputWide{}.Map(e))
	}

	return out
}

func (do DevicesConsoleOutput) Sort() DevicesConsoleOutput {
	sort.Slice(do, func(i, j int) bool {
		return do[j].ID > do[i].ID
	})

	return do
}

func (do DevicesConsoleOutputWide) Sort() DevicesConsoleOutputWide {
	sort.Slice(do, func(i, j int) bool {
		return do[j].ID > do[i].ID
	})

	return do
}
