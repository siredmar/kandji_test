package printer

import (
	"fmt"
	"time"

	api "github.com/grid-x/gxctl/pkg/api"
)

type PodConsoleOutput struct {
	ID       string `header:"ID"`
	DeviceID string `header:"Device ID"`
	Age      string `header:"Age"`
}

func (o PodConsoleOutput) Map(p api.Pod) PodConsoleOutput {
	o.ID = p.Metadata.ID
	o.DeviceID = p.Spec.DeviceID
	if p.Status.StartTime != nil {
		startTime := p.Status.StartTime

		if time.Since(startTime.Time).Seconds() < 60 {
			o.Age = fmt.Sprintf("%.0fs", time.Since(startTime.Time).Seconds())
		} else if time.Since(startTime.Time).Minutes() < 60 {
			o.Age = fmt.Sprintf("%.0fm", time.Since(startTime.Time).Minutes())
		} else if time.Since(startTime.Time).Hours() < 24 {
			o.Age = fmt.Sprintf("%.0fh", time.Since(startTime.Time).Hours())
		} else {
			o.Age = fmt.Sprintf("%.0fd", time.Since(startTime.Time).Hours()/24)
		}
	}
	return o
}
