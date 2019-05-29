package printer

import (
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
		o.Selector = SortedString(d.Spec.Selector.MatchByLabels)
	}

	return o
}
