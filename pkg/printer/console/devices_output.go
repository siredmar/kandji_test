package printer

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/PaesslerAG/jsonpath"
	api "github.com/grid-x/gxctl/pkg/api"
)

type DevicesConsoleOutput struct {
	raw api.Devices
}
type DevicesConsoleOutputWide struct {
	raw api.Devices
}

func (do DevicesConsoleOutput) Inject(i api.Devices) DevicesConsoleOutput {
	do.raw = i

	return do
}

func (do DevicesConsoleOutputWide) Inject(i api.Devices) DevicesConsoleOutputWide {
	do.raw = i

	return do
}

func (do DevicesConsoleOutput) Filter(showAll bool) DevicesConsoleOutput {
	var r []api.Device
	for _, e := range do.raw.Devices {
		if !showAll && e.Status.LastHeartbeat == nil {
			continue
		}
		r = append(r, e)
	}
	do.raw.Devices = r

	return do
}

func (do DevicesConsoleOutputWide) Filter(showAll bool) DevicesConsoleOutputWide {
	var r []api.Device
	for _, e := range do.raw.Devices {
		if !showAll && e.Status.LastHeartbeat == nil {
			continue
		}
		r = append(r, e)
	}
	do.raw.Devices = r

	return do
}

func (do DevicesConsoleOutput) Map() []DeviceConsoleOutput {
	var output []DeviceConsoleOutput
	for _, e := range do.raw.Devices {
		output = append(output, DeviceConsoleOutput{}.Map(e))
	}

	return output
}

func (do DevicesConsoleOutputWide) Map() []DeviceConsoleOutputWide {
	var output []DeviceConsoleOutputWide
	for _, e := range do.raw.Devices {
		output = append(output, DeviceConsoleOutputWide{}.Map(e))
	}

	return output
}

func (do DevicesConsoleOutput) Sort(sortBy string) DevicesConsoleOutput {
	s := interface{}(nil)
	rawJson, _ := json.Marshal(do.raw)
	json.Unmarshal(rawJson, &s)

	var found bool
	sort.Slice(do.raw.Devices, func(i, j int) bool {
		di, _ := jsonpath.Get(fmt.Sprintf("$..devices[?(@.metadata.id==\"%s\")].%s", do.raw.Devices[i].Metadata.ID, sortBy), s)
		dj, _ := jsonpath.Get(fmt.Sprintf("$..devices[?(@.metadata.id==\"%s\")].%s", do.raw.Devices[j].Metadata.ID, sortBy), s)

		if fmt.Sprintf("%v", di) != "[]" {
			found = true
		}
		return fmt.Sprintf("%v", di) < fmt.Sprintf("%v", dj)
	})

	if !found {
		s := fmt.Sprintf("Warning: Did not find jsonpath value: %s. Result is unsorted", sortBy)
		fmt.Println(s)
	}

	return do
}

func (do DevicesConsoleOutputWide) Sort(sortBy string) DevicesConsoleOutputWide {
	s := interface{}(nil)
	rawJson, _ := json.Marshal(do.raw)
	json.Unmarshal(rawJson, &s)

	sort.Slice(do.raw.Devices, func(i, j int) bool {
		di, _ := jsonpath.Get(fmt.Sprintf("$..devices[?(@.metadata.id==\"%s\")].%s", do.raw.Devices[i].Metadata.ID, sortBy), s)
		dj, _ := jsonpath.Get(fmt.Sprintf("$..devices[?(@.metadata.id==\"%s\")].%s", do.raw.Devices[j].Metadata.ID, sortBy), s)

		return fmt.Sprintf("%v", di) < fmt.Sprintf("%v", dj)
	})

	return do
}
