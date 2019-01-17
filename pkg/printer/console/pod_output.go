package printer

import (
	api "github.com/grid-x/gxctl/pkg/api"
)

type PodConsoleOutput struct {
	ID        string `header:"ID"`
	DeviceID  string `header:"Device ID"`
	StartTime string `header:"StartTime"`
}

func (o PodConsoleOutput) Map(p api.Pod) PodConsoleOutput {
	o.ID = p.UUID
	o.DeviceID = p.Spec.DeviceID
	if p.Status.StartTime != nil {
		o.StartTime = p.Status.StartTime.String()
	}
	return o
}
