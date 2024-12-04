package printer

import (
	api "github.com/grid-x/gxctl/pkg/api"
)

type DevicesLogsConsoleOutput []DeviceLogsConsoleOutput

func (o DevicesLogsConsoleOutput) Map(v api.DevicesLogs) DevicesLogsConsoleOutput {
	o = make(DevicesLogsConsoleOutput, len(v.Items))

	for i, l := range v.Items {
		o[i] = DeviceLogsConsoleOutput{}.Map(api.DeviceLogs(l))
	}

	return o
}

type DevicesLogsConsoleOutputWide []DeviceLogsConsoleOutputWide

func (o DevicesLogsConsoleOutputWide) Map(v api.DevicesLogs) DevicesLogsConsoleOutputWide {
	o = make(DevicesLogsConsoleOutputWide, len(v.Items))

	for i, l := range v.Items {
		o[i] = DeviceLogsConsoleOutputWide{}.Map(api.DeviceLogs(l))
	}

	return o
}
