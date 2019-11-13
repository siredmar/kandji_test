package printer

import (
	"strconv"
	"time"

	api "github.com/grid-x/gxctl/pkg/api"
	units "github.com/grid-x/gxctl/pkg/printer/units"
)

type MaintenanceConsoleOutput struct {
	DeviceID   string `header:"DeviceID"`
	Successful string `header:"Successful"`
	Failed     string `header:"Failed"`
	Running    string `header:"Running"`
}

type MaintenanceConsoleOutputWide struct {
	DeviceID   string `header:"ID"`
	StartedAt  string `header:"StartedAt"`
	FinishedAt string `header:"FinishedAt"`
	Successful string `header:"Successful"`
	Failed     string `header:"Failed"`
	Running    string `header:"Running"`
}

func (o MaintenanceConsoleOutput) Map(m api.MaintenanceTask) MaintenanceConsoleOutput {
	o.DeviceID = m.Spec.DeviceID
	o.Successful = strconv.Itoa(m.Status.Successful)
	o.Failed = strconv.Itoa(m.Status.Failed)
	o.Running = strconv.Itoa(m.Status.Running)

	return o
}

func (o MaintenanceConsoleOutputWide) Map(m api.MaintenanceTask) MaintenanceConsoleOutputWide {
	o.DeviceID = m.Spec.DeviceID
	o.Successful = strconv.Itoa(m.Status.Successful)
	o.Failed = strconv.Itoa(m.Status.Failed)
	o.Running = strconv.Itoa(m.Status.Running)

	if m.Status.StartedAt != nil {
		o.StartedAt = units.HumanDuration(time.Now().Sub(m.Status.StartedAt.Time)) + " ago"
	}

	if m.Status.FinishedAt != nil {
		o.FinishedAt = units.HumanDuration(time.Now().Sub(m.Status.FinishedAt.Time)) + " ago"
	}

	return o
}
