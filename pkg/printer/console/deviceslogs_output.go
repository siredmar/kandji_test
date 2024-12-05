package printer

import (
	api "github.com/grid-x/gxctl/pkg/api"
)

type DevicesLogsConsoleOutput []DeviceLogsConsoleOutput

func (o DevicesLogsConsoleOutput) Map(v api.DevicesLogsWithSerialNumber) DevicesLogsConsoleOutput {
	o = make(DevicesLogsConsoleOutput, len(v))

	for i, l := range v {
		o[i] = DeviceLogsConsoleOutput{}.Map(l)
	}

	return o
}

type DevicesLogsConsoleOutputWide []DeviceLogsConsoleOutputWide

func (o DevicesLogsConsoleOutputWide) Map(v api.DevicesLogsWithSerialNumber) DevicesLogsConsoleOutputWide {
	o = make(DevicesLogsConsoleOutputWide, len(v))

	for i, l := range v {
		o[i] = DeviceLogsConsoleOutputWide{}.Map(l)
	}

	return o
}
