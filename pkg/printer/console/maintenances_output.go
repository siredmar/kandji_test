package printer

import (
	"sort"

	api "github.com/grid-x/gxctl/pkg/api"
)

type MaintenancesConsoleOutput []MaintenanceConsoleOutput

func (mo MaintenancesConsoleOutput) Map(m api.MaintenanceTasks) MaintenancesConsoleOutput {
	var out MaintenancesConsoleOutput
	for _, e := range m.MaintenanceTasks {
		out = append(out, MaintenanceConsoleOutput{}.Map(e))
	}

	return out
}

func (mo MaintenancesConsoleOutput) Sort() MaintenancesConsoleOutput {
	sort.Slice(mo, func(i, j int) bool {
		return mo[j].ID > mo[i].ID
	})

	return mo
}
