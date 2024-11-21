package printer

import (
	api "github.com/grid-x/gxctl/pkg/api"
)

type DeviceConfigMapsConsoleOutput struct {
	raw api.DeviceConfigMaps
}
type DeviceConfigMapsConsoleOutputWide struct {
	raw api.DeviceConfigMaps
}

func (do DeviceConfigMapsConsoleOutput) Inject(i api.DeviceConfigMaps) DeviceConfigMapsConsoleOutput {
	do.raw = i

	return do
}

func (do DeviceConfigMapsConsoleOutputWide) Inject(i api.DeviceConfigMaps) DeviceConfigMapsConsoleOutputWide {
	do.raw = i

	return do
}

func (do DeviceConfigMapsConsoleOutput) Map() []DeviceConfigMapConsoleOutput {
	output := make([]DeviceConfigMapConsoleOutput, len(do.raw.DeviceConfigMaps))
	for _, e := range do.raw.DeviceConfigMaps {
		output = append(output, DeviceConfigMapConsoleOutput{}.Map(e))
	}

	return output
}

func (do DeviceConfigMapsConsoleOutputWide) Map() []DeviceConfigMapConsoleOutputWide {
	output := make([]DeviceConfigMapConsoleOutputWide, len(do.raw.DeviceConfigMaps))
	for _, e := range do.raw.DeviceConfigMaps {
		output = append(output, DeviceConfigMapConsoleOutputWide{}.Map(e)...)
		output = append(output, DeviceConfigMapConsoleOutputWide{})
	}

	return output
}
