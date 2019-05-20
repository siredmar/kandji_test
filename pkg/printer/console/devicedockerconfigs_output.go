package printer

import (
	"sort"

	api "github.com/grid-x/gxctl/pkg/api"
)

type DeviceDockerConfigsConsoleOutput []DeviceDockerConfigConsoleOutput
type DeviceDockerConfigsConsoleOutputWide []DeviceDockerConfigConsoleOutputWide

func (do DeviceDockerConfigsConsoleOutput) Map(d api.DeviceDockerConfigs) DeviceDockerConfigsConsoleOutput {
	var out DeviceDockerConfigsConsoleOutput
	for _, e := range d.DeviceDockerConfigs {
		out = append(out, DeviceDockerConfigConsoleOutput{}.Map(e))
	}

	return out
}

func (do DeviceDockerConfigsConsoleOutputWide) Map(d api.DeviceDockerConfigs) DeviceDockerConfigsConsoleOutputWide {
	var out DeviceDockerConfigsConsoleOutputWide
	for _, e := range d.DeviceDockerConfigs {
		out = append(out, DeviceDockerConfigConsoleOutputWide{}.Map(e))
	}

	return out
}

func (do DeviceDockerConfigsConsoleOutput) Sort() DeviceDockerConfigsConsoleOutput {
	sort.Slice(do, func(i, j int) bool {
		return do[j].ID > do[i].ID
	})

	return do
}

func (do DeviceDockerConfigsConsoleOutputWide) Sort() DeviceDockerConfigsConsoleOutputWide {
	sort.Slice(do, func(i, j int) bool {
		return do[j].ID > do[i].ID
	})

	return do
}
