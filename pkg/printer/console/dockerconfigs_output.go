package printer

import (
	"sort"

	api "github.com/grid-x/gxctl/pkg/api"
)

type DockerConfigsConsoleOutput []DockerConfigConsoleOutput
type DockerConfigsConsoleOutputWide []DockerConfigConsoleOutputWide

func (do DockerConfigsConsoleOutput) Map(d api.DockerConfigs) DockerConfigsConsoleOutput {
	var out DockerConfigsConsoleOutput
	for _, e := range d.DockerConfigs {
		out = append(out, DockerConfigConsoleOutput{}.Map(e))
	}

	return out
}

func (do DockerConfigsConsoleOutputWide) Map(d api.DockerConfigs) DockerConfigsConsoleOutputWide {
	var out DockerConfigsConsoleOutputWide
	for _, e := range d.DockerConfigs {
		out = append(out, DockerConfigConsoleOutputWide{}.Map(e))
	}

	return out
}

func (do DockerConfigsConsoleOutput) Sort() DockerConfigsConsoleOutput {
	sort.Slice(do, func(i, j int) bool {
		return do[j].ID > do[i].ID
	})

	return do
}

func (do DockerConfigsConsoleOutputWide) Sort() DockerConfigsConsoleOutputWide {
	sort.Slice(do, func(i, j int) bool {
		return do[j].ID > do[i].ID
	})

	return do
}
