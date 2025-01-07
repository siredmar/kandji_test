package printer

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/PaesslerAG/jsonpath"

	api "github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/filter"
)

type DevicesConsoleOutput struct {
	raw api.Devices
}
type DevicesConsoleOutputWide struct {
	raw api.Devices
}

func (do DevicesConsoleOutput) Filter(f filter.Filter) DevicesConsoleOutput {
	do.raw = filterDevices(do.raw, f)

	return do
}

func (do DevicesConsoleOutputWide) Filter(f filter.Filter) DevicesConsoleOutputWide {
	do.raw = filterDevices(do.raw, f)

	return do
}

func filterDevices(devices api.Devices, f filter.Filter) api.Devices {
	if f == nil {
		return devices
	}

	var out api.Devices

	for _, d := range devices.Devices {
		include, err := f.Eval(d)
		if err != nil || !include {
			continue
		}
		out.Devices = append(out.Devices, d)
	}

	return out
}

func (do DevicesConsoleOutput) Inject(i api.Devices) DevicesConsoleOutput {
	do.raw = i

	return do
}

func (do DevicesConsoleOutputWide) Inject(i api.Devices) DevicesConsoleOutputWide {
	do.raw = i

	return do
}

func (do DevicesConsoleOutput) Map() []DeviceConsoleOutput {
	output := []DeviceConsoleOutput{}
	for _, e := range do.raw.Devices {
		output = append(output, DeviceConsoleOutput{}.Map(e))
	}

	return output
}

func (do DevicesConsoleOutputWide) Map() []DeviceConsoleOutputWide {
	output := []DeviceConsoleOutputWide{}
	for _, e := range do.raw.Devices {
		output = append(output, DeviceConsoleOutputWide{}.Map(e))
	}

	return output
}
func (do DevicesConsoleOutput) Sort(sortBy string) DevicesConsoleOutput {
	out := sortDevices(do.raw, sortBy)
	do.raw.Devices = out

	return do
}

func (do DevicesConsoleOutputWide) Sort(sortBy string) DevicesConsoleOutputWide {
	out := sortDevices(do.raw, sortBy)
	do.raw.Devices = out

	return do
}

func sortDevices(in api.Devices, sortBy string) []api.Device {
	s := interface{}(nil)
	rawJSON, _ := json.Marshal(in)
	_ = json.Unmarshal(rawJSON, &s)

	if len(in.Devices) < 2 {
		return in.Devices
	}

	sortValues := make(map[string]interface{})

	for _, d := range in.Devices {
		sortValues[d.Metadata.ID], _ = jsonpath.Get(fmt.Sprintf("$..devices[?(@.metadata.id==\"%s\")].%s", d.Metadata.ID, sortBy), s)
	}

	var found bool
	sort.Slice(in.Devices, func(i, j int) bool {
		di := sortValues[in.Devices[i].Metadata.ID]
		dj := sortValues[in.Devices[j].Metadata.ID]

		if fmt.Sprintf("%v", di) != "[]" {
			found = true
		}
		return fmt.Sprintf("%v", di) < fmt.Sprintf("%v", dj)
	})

	if !found {
		s := fmt.Sprintf("Warning: Did not find jsonpath value: %s. Result is unsorted", sortBy)
		fmt.Println(s)
	}

	return in.Devices
}
