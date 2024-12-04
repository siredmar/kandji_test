package printer

import (
	api "github.com/grid-x/gxctl/pkg/api"
)

type DevicesLogsConsoleOutputWide []DeviceLogsConsoleOutputWide

func (o DevicesLogsConsoleOutputWide) Map(v api.DevicesLogs) DevicesLogsConsoleOutputWide {
	o = make(DevicesLogsConsoleOutputWide, len(v.Items))

	for i, l := range v.Items {
		o[i] = DeviceLogsConsoleOutputWide{}.Map(api.DeviceLogs(l))
	}

	return o
}
