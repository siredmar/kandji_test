package printer

import (
	"time"

	api "github.com/grid-x/gxctl/pkg/api"
)

type DeviceLogsConsoleOutput struct {
	SerialNumber string `header:"serial_number"`
	ExpiresAt    string `header:"expires_at"`
	LogLevel     string `header:"log_level"`
	NotifiedAt   string `header:"notified_at"`
}

func (o DeviceLogsConsoleOutput) Map(v api.DeviceLogsWithSerialNumber) DeviceLogsConsoleOutput {
	o.SerialNumber = v.SerialNumber
	o.ExpiresAt = v.Spec.ExpiresAt.Format(time.RFC3339)
	o.LogLevel = v.Spec.LogLevel
	if len(v.Status.NotifiedAt) > 0 {
		o.NotifiedAt = v.Status.NotifiedAt[0].Format(time.RFC3339)
	}

	return o
}

type DeviceLogsConsoleOutputWide struct {
	SerialNumber string `header:"serial_number"`
	ExpiresAt    string `header:"expires_at"`
	LogLevel     string `header:"log_level"`
	NotifiedAt   string `header:"notified_at"`
	Owner        string `header:"owner"`
	ID           string `header:"id"`
}

func (o DeviceLogsConsoleOutputWide) Map(v api.DeviceLogsWithSerialNumber) DeviceLogsConsoleOutputWide {
	o.SerialNumber = v.SerialNumber
	o.ExpiresAt = v.Spec.ExpiresAt.Format(time.RFC3339)
	o.LogLevel = v.Spec.LogLevel
	if len(v.Status.NotifiedAt) > 0 {
		o.NotifiedAt = v.Status.NotifiedAt[0].Format(time.RFC3339)
	}

	o.Owner = v.Spec.Owner
	o.ID = v.Metadata.ID

	return o
}
