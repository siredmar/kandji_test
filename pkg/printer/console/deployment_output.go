package printer

import (
	"fmt"

	api "github.com/grid-x/gxctl/pkg/api"
)

type DeploymentConsoleOutput struct {
	ID       string `header:"ID"`
	App      string `header:"Application"`
	Selector string `header:"Sector"`
	Images   string `header:"Images"`
}

func (o DeploymentConsoleOutput) Map(d api.Deployment) DeploymentConsoleOutput {
	o.ID = d.Metadata.ID
	o.App = d.Spec.App

	if d.Spec.Selector.MatchByLabels != nil {
		var s string
		for key, value := range d.Spec.Selector.MatchByLabels {
			s += fmt.Sprintf("%s:%s\n", key, value)
		}

		o.Selector = s[:len(s)-1]
	}

	if d.Spec.Template.Spec.Containers != nil {
		var s string
		for _, value := range d.Spec.Template.Spec.Containers {
			s += fmt.Sprintf("%s\n", value.Image)
		}

		o.Images = s[:len(s)-1]
	}

	return o
}
