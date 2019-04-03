package printer

import (
	api "github.com/grid-x/gxctl/pkg/api"
)

type DockerConfigConsoleOutput struct {
	ID       string `header:"ID"`
	Registry string `header:"Registry"`
}

type DockerConfigConsoleOutputWide struct {
	ID       string `header:"ID"`
	Registry string `header:"Registry"`
}

func (o DockerConfigConsoleOutput) Map(d api.DockerConfig) DockerConfigConsoleOutput {
	o.ID = d.Metadata.ID
	o.Registry = d.Spec.Registry

	return o
}

func (o DockerConfigConsoleOutputWide) Map(d api.DockerConfig) DockerConfigConsoleOutputWide {
	o.ID = d.Metadata.ID
	o.Registry = d.Spec.Registry

	return o
}
