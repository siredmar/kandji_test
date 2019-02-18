package printer

import (
	"fmt"
	"strconv"

	api "github.com/grid-x/gxctl/pkg/api"
)

type MaintenanceConsoleOutput struct {
	ID         string `header:"ID"`
	Successful string `header:"Successful"`
	Failed     string `header:"Failed"`
	Running    string `header:"Running"`
}

type MaintenanceConsoleOutputWide struct {
	ID         string `header:"ID"`
	Successful string `header:"Successful"`
	Failed     string `header:"Failed"`
	Running    string `header:"Running"`
	Selector   string `header:"Selector"`
}

func (o MaintenanceConsoleOutput) Map(m api.MaintenanceTask) MaintenanceConsoleOutput {
	o.ID = m.Metadata.ID
	o.Successful = strconv.Itoa(m.Status.Successful)
	o.Failed = strconv.Itoa(m.Status.Failed)
	o.Running = strconv.Itoa(m.Status.Running)

	return o
}

func (o MaintenanceConsoleOutputWide) Map(m api.MaintenanceTask) MaintenanceConsoleOutputWide {
	o.ID = m.Metadata.ID
	o.Successful = strconv.Itoa(m.Status.Successful)
	o.Failed = strconv.Itoa(m.Status.Failed)
	o.Running = strconv.Itoa(m.Status.Running)

	if m.Spec.Selector.MatchByLabels != nil {
		var s string
		for key, value := range m.Spec.Selector.MatchByLabels {
			s += fmt.Sprintf("%s:%s\n", key, value)
		}

		o.Selector = s[:len(s)-1]
	}

	return o
}
