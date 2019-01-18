package printer

import (
	api "github.com/grid-x/gxctl/pkg/api"
)

type ApplicationConsoleOutput struct {
	Name string `header:"Name"`
}

func (o ApplicationConsoleOutput) Map(a api.Application) ApplicationConsoleOutput {
	o.Name = a.Name

	return o
}
