package printer

import (
	"sort"

	api "github.com/grid-x/gxctl/pkg/api"
)

type DevicesConsoleOutput []DeviceConsoleOutput
type DevicesConsoleOutputWide []DeviceConsoleOutputWide

func (do DevicesConsoleOutput) Map(d api.Devices, showAll bool) DevicesConsoleOutput {
	var out DevicesConsoleOutput
	for _, e := range d.Devices {
		if !showAll && e.Status.LastHeartbeat == nil {
			continue
		}
		out = append(out, DeviceConsoleOutput{}.Map(e))
	}

	return out
}

func (do DevicesConsoleOutputWide) Map(d api.Devices, showAll bool) DevicesConsoleOutputWide {
	var out DevicesConsoleOutputWide
	for _, e := range d.Devices {
		if !showAll && e.Status.LastHeartbeat == nil {
			continue
		}
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
