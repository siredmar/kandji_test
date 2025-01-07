package printer

import (
	"sort"

	api "github.com/grid-x/gxctl/pkg/api"
)

type ApplicationsConsoleOutput []ApplicationConsoleOutput

func (ApplicationsConsoleOutput) Map(a api.Applications) ApplicationsConsoleOutput {
	out := ApplicationsConsoleOutput{}
	for _, e := range a.Applications {
		out = append(out, ApplicationConsoleOutput{}.Map(e))
	}

	return out
}

func (ao ApplicationsConsoleOutput) Sort() ApplicationsConsoleOutput {
	sort.Slice(ao, func(i, j int) bool {
		return ao[j].Name > ao[i].Name
	})

	return ao
}
