package printer

import (
	"sort"

	api "github.com/grid-x/gxctl/pkg/api"
)

type MaintenancesConsoleOutput []MaintenanceConsoleOutput
type MaintenancesConsoleOutputWide []MaintenanceConsoleOutputWide

func (mo MaintenancesConsoleOutput) Map(m api.MaintenanceTasks) MaintenancesConsoleOutput {
	output := MaintenancesConsoleOutput{}
	for _, e := range m.MaintenanceTasks {
		output = append(output, MaintenanceConsoleOutput{}.Map(e))
	}

	return output
}

func (mo MaintenancesConsoleOutputWide) Map(m api.MaintenanceTasks) MaintenancesConsoleOutputWide {
	output := MaintenancesConsoleOutputWide{}
	for _, e := range m.MaintenanceTasks {
		output = append(output, MaintenanceConsoleOutputWide{}.Map(e))
	}

	return output
}

func (mo MaintenancesConsoleOutput) Sort() MaintenancesConsoleOutput {
	sort.Slice(mo, func(i, j int) bool {
		return mo[j].DeviceID > mo[i].DeviceID
	})

	return mo
}

func (mo MaintenancesConsoleOutputWide) Sort() MaintenancesConsoleOutputWide {
	sort.Slice(mo, func(i, j int) bool {
		return mo[j].DeviceID > mo[i].DeviceID
	})

	return mo
}
