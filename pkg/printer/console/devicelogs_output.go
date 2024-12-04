package printer

import (
	"time"

	api "github.com/grid-x/gxctl/pkg/api"
)

type DeviceLogsConsoleOutputWide struct {
	ID         string `header:"id"`
	ExpiresAt  string `header:"expires_at"`
	LogLevel   string `header:"log_level"`
	Owner      string `header:"owner"`
	NotifiedAt string `header:"notified_at"`
}

func (o DeviceLogsConsoleOutputWide) Map(v api.DeviceLogs) DeviceLogsConsoleOutputWide {
	o.ID = v.Metadata.ID
	o.ExpiresAt = v.Spec.ExpiresAt.Format(time.RFC3339)
	o.LogLevel = v.Spec.LogLevel
	o.Owner = v.Spec.Owner
	if len(v.Status.NotifiedAt) > 0 {
		o.NotifiedAt = v.Status.NotifiedAt[0].Format(time.RFC3339)
	}

	return o
}
