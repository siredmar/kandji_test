package printer

import (
	"fmt"
	"time"

	api "github.com/grid-x/gxctl/pkg/api"
	units "github.com/grid-x/gxctl/pkg/printer/units"
)

type PodConsoleOutput struct {
	ID       string `header:"ID"`
	DeviceID string `header:"Device ID"`
	Age      string `header:"Age"`
	Image    string `header:"Images"`
}

type PodConsoleOutputWide struct {
	ID         string `header:"ID"`
	DeviceID   string `header:"Device ID"`
	Age        string `header:"Age"`
	Containers string `header:"Containers"`
	Image      string `header:"Images"`
}

func (o PodConsoleOutput) Map(p api.Pod) PodConsoleOutput {
	o.ID = p.Metadata.ID
	o.DeviceID = p.Spec.DeviceID

	var images string
	for _, container := range p.Spec.Config.Containers {
		images += fmt.Sprintf("%s \n", container.Image)
	}
	o.Image = images

	if p.Status.StartTime != nil {
		o.Age = units.HumanDuration(time.Now().Sub(p.Status.StartTime.Time)) + " ago"
	}
	return o
}

func (o PodConsoleOutputWide) Map(p api.Pod) PodConsoleOutputWide {
	o.ID = p.Metadata.ID
	o.DeviceID = p.Spec.DeviceID
	var images string
	var containers string
	for _, container := range p.Spec.Config.Containers {
		images += fmt.Sprintf("%s \n", container.Image)
		containers += fmt.Sprintf("%s \n", container.Name)
	}
	o.Image = images
	o.Containers = containers

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
