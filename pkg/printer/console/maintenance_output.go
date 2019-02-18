package printer

import (
	api "github.com/grid-x/gxctl/pkg/api"
	"strconv"
)

type MaintenanceConsoleOutput struct {
	ID         string `header:"ID"`
	Successful string `header:"Successful"`
	Failed     string `header:"Failed"`
	Running    string `header:"Running"`
}

func (o MaintenanceConsoleOutput) Map(m api.MaintenanceTask) MaintenanceConsoleOutput {
	o.ID = m.Metadata.ID
	o.Successful = strconv.Itoa(m.Status.Successful)
	o.Failed = strconv.Itoa(m.Status.Failed)
	o.Running = strconv.Itoa(m.Status.Running)

	return o
}
