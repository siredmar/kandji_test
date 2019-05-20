package printer

import (
	api "github.com/grid-x/gxctl/pkg/api"
)

type DeviceDockerConfigConsoleOutput struct {
	ID       string `header:"ID"`
	Type     string `header:"Type"`
	Registry string `header:"Registry"`
}

type DeviceDockerConfigConsoleOutputWide struct {
	ID       string `header:"ID"`
	Type     string `header:"Type"`
	Registry string `header:"Registry"`
	Selector string `header:"Selector"`
}

func (o DeviceDockerConfigConsoleOutput) Map(d api.DeviceDockerConfig) DeviceDockerConfigConsoleOutput {
	o.ID = d.Metadata.ID
	o.Registry = d.Spec.Registry

	if d.Spec.Credentials.AWS != nil {
		o.Type = "AWS"
	} else if d.Spec.Credentials.DockerHub != nil {
		o.Type = "DockerHub"
	}

	return o
}

func (o DeviceDockerConfigConsoleOutputWide) Map(d api.DeviceDockerConfig) DeviceDockerConfigConsoleOutputWide {
	o.ID = d.Metadata.ID
	o.Registry = d.Spec.Registry

	if d.Spec.Credentials.AWS != nil {
		o.Type = "AWS"
	} else if d.Spec.Credentials.DockerHub != nil {
		o.Type = "DockerHub"
	}

	return o
}
