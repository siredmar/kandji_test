package printer

import (
	"sort"

	api "github.com/grid-x/gxctl/pkg/api"
)

type MaintenancesConsoleOutput []MaintenanceConsoleOutput
type MaintenancesConsoleOutputWide []MaintenanceConsoleOutputWide

func (mo MaintenancesConsoleOutput) Map(m api.MaintenanceTasks) MaintenancesConsoleOutput {
	var out MaintenancesConsoleOutput
	for _, e := range m.MaintenanceTasks {
		out = append(out, MaintenanceConsoleOutput{}.Map(e))
	}

	return out
}

func (mo MaintenancesConsoleOutputWide) Map(m api.MaintenanceTasks) MaintenancesConsoleOutputWide {
	var out MaintenancesConsoleOutputWide
	for _, e := range m.MaintenanceTasks {
		out = append(out, MaintenanceConsoleOutputWide{}.Map(e))
	}

	return out
}

func (mo MaintenancesConsoleOutput) Sort() MaintenancesConsoleOutput {
	sort.Slice(mo, func(i, j int) bool {
		return mo[j].ID > mo[i].ID
	})

	return mo
}

func (mo MaintenancesConsoleOutputWide) Sort() MaintenancesConsoleOutputWide {
	sort.Slice(mo, func(i, j int) bool {
		return mo[j].ID > mo[i].ID
	})

	return mo
}
