package printer

import (
	"sort"

	api "github.com/grid-x/gxctl/pkg/api"
)

type DeploymentsConsoleOutput []DeploymentConsoleOutput

func (do DeploymentsConsoleOutput) Map(d api.Deployments) DeploymentsConsoleOutput {
	var out DeploymentsConsoleOutput
	for _, e := range d.Deployments {
		out = append(out, DeploymentConsoleOutput{}.Map(e))
	}

	return out
}

func (do DeploymentsConsoleOutput) Sort() DeploymentsConsoleOutput {
	sort.Slice(do, func(i, j int) bool {
		return do[j].App > do[i].App
	})

	return do
}
