package printer

import (
	"sort"

	api "github.com/grid-x/gxctl/pkg/api"
)

type PodsConsoleOutput []PodConsoleOutput

func (po PodsConsoleOutput) Map(p api.Pods) PodsConsoleOutput {
	var out PodsConsoleOutput
	for _, e := range p.Pods {
		out = append(out, PodConsoleOutput{}.Map(e))
	}

	return out
}

func (po PodsConsoleOutput) Sort() PodsConsoleOutput {
	sort.Slice(po, func(i, j int) bool {
		return po[j].ID > po[i].ID
	})

	return po
}
