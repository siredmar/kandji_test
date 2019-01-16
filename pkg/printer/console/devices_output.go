package printer

import (
	"github.com/landoop/tableprinter"
	"os"
	"sort"

	api "github.com/grid-x/gxctl/pkg/api"
)

type DevicesConsoleOutput []DeviceConsoleOutput

func (do DevicesConsoleOutput) Print() int {
	printer := tableprinter.New(os.Stdout)
	printer.HeaderLine = false
	return printer.Print(do)
}

func (do DevicesConsoleOutput) Map(d api.Devices) DevicesConsoleOutput {
	var out DevicesConsoleOutput
	for _, e := range d.Devices {
		out = append(out, DeviceConsoleOutput{}.Map(e))
	}

	return out
}

func (do DevicesConsoleOutput) Sort() DevicesConsoleOutput {
	sort.Slice(do, func(i, j int) bool {
		return do[j].ID > do[i].ID
	})

	return do
}
