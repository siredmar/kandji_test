package printer

import (
	"fmt"

	api "github.com/grid-x/gxctl/pkg/api"
)

type DockerConfigConsoleOutput struct {
	ID       string `header:"ID"`
	Type     string `header:"Type"`
	Registry string `header:"Registry"`
}

type DockerConfigConsoleOutputWide struct {
	ID       string `header:"ID"`
	Type     string `header:"Type"`
	Registry string `header:"Registry"`
	Selector string `header:"Selector"`
}

func (o DockerConfigConsoleOutput) Map(d api.DockerConfig) DockerConfigConsoleOutput {
	o.ID = d.Metadata.ID
	o.Registry = d.Spec.Registry

	if d.Spec.Credentials.AWS != nil {
		o.Type = "AWS"
	} else if d.Spec.Credentials.DockerHub != nil {
		o.Type = "DockerHub"
	}

	return o
}

func (o DockerConfigConsoleOutputWide) Map(d api.DockerConfig) DockerConfigConsoleOutputWide {
	o.ID = d.Metadata.ID
	o.Registry = d.Spec.Registry

	if d.Spec.Credentials.AWS != nil {
		o.Type = "AWS"
	} else if d.Spec.Credentials.DockerHub != nil {
		o.Type = "DockerHub"
	}

	if d.Spec.Selector.MatchByLabels != nil {
		var s string
		for key, value := range d.Spec.Selector.MatchByLabels {
			s += fmt.Sprintf("%s:%s\n", key, value)
		}

		o.Selector = s[:len(s)-1]
	}

	return o
}
