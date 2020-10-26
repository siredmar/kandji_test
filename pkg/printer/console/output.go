package printer

import (
	"errors"
	"fmt"
	"io"
	"sort"

	"github.com/landoop/tableprinter"

	api "github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/printer/filter"
)

type ConsolePrinter struct {
	Target io.Writer
}

type ConsolePrintConfig struct {
	Filter filter.Filter
	SortBy  string
	ShowAll bool
}

func NewConsolePrinter(t io.Writer) *ConsolePrinter {
	return &ConsolePrinter{
		Target: t,
	}
}

func (c *ConsolePrinter) Print(v interface{}, config ConsolePrintConfig) error {
	var out interface{}

	switch v := v.(type) {
	case api.Device:
		out = DeviceConsoleOutput{}.Map(v)
	case api.Devices:
		out = DevicesConsoleOutput{}.Inject(v).ShowAll(config.ShowAll).Filter(config.Filter).Sort(config.SortBy).Map()
	case api.Pod:
		out = PodConsoleOutput{}.Map(v)
	case api.Pods:
		out = PodsConsoleOutput{}.Inject(v).ShowAll(config.ShowAll).Sort(config.SortBy).Map()
	case api.Deployment:
		out = DeploymentConsoleOutput{}.Map(v)
	case api.Deployments:
		out = DeploymentsConsoleOutput{}.Inject(v).ShowAll(config.ShowAll).Sort(config.SortBy).Map()
	case api.Application:
		out = ApplicationConsoleOutput{}.Map(v)
	case api.Applications:
		out = ApplicationsConsoleOutput{}.Map(v).Sort()
	case api.MaintenanceTask:
		out = MaintenanceConsoleOutput{}.Map(v)
	case api.MaintenanceTasks:
		out = MaintenancesConsoleOutput{}.Map(v).Sort()
	case api.DockerConfig:
		out = DockerConfigConsoleOutput{}.Map(v)
	case api.DockerConfigs:
		out = DockerConfigsConsoleOutput{}.Map(v).Sort()
	case api.DeviceDockerConfig:
		out = DeviceDockerConfigConsoleOutput{}.Map(v)
	case api.DeviceDockerConfigs:
		out = DeviceDockerConfigsConsoleOutput{}.Map(v).Sort()
	default:
		s := fmt.Sprintf("Not able to print to console! Unknow type %s.", v)
		return errors.New(s)
	}

	printer := tableprinter.New(c.Target)
	printer.HeaderLine = false

	printer.Print(out)
	return nil
}

func (c *ConsolePrinter) PrintWide(v interface{}, config ConsolePrintConfig) error {
	var out interface{}

	switch v := v.(type) {
	case api.Device:
		out = DeviceConsoleOutputWide{}.Map(v)
	case api.Devices:
		out = DevicesConsoleOutputWide{}.Inject(v).ShowAll(config.ShowAll).Filter(config.Filter).Sort(config.SortBy).Map()
	case api.Pod:
		out = PodConsoleOutputWide{}.Map(v)
	case api.Pods:
		out = PodsConsoleOutputWide{}.Inject(v).ShowAll(config.ShowAll).Sort(config.SortBy).Map()
	case api.Deployment:
		out = DeploymentConsoleOutputWide{}.Map(v)
	case api.Deployments:
		out = DeploymentsConsoleOutputWide{}.Inject(v).ShowAll(config.ShowAll).Sort(config.SortBy).Map()
	case api.Application:
		out = ApplicationConsoleOutput{}.Map(v) //TODO make wide mapping
	case api.Applications:
		out = ApplicationsConsoleOutput{}.Map(v).Sort() //TODO make wide mapping
	case api.MaintenanceTask:
		out = MaintenanceConsoleOutputWide{}.Map(v)
	case api.MaintenanceTasks:
		out = MaintenancesConsoleOutputWide{}.Map(v).Sort()
	case api.DockerConfig:
		out = DockerConfigConsoleOutputWide{}.Map(v)
	case api.DockerConfigs:
		out = DockerConfigsConsoleOutputWide{}.Map(v).Sort()
	case api.DeviceDockerConfig:
		out = DeviceDockerConfigConsoleOutputWide{}.Map(v)
	case api.DeviceDockerConfigs:
		out = DeviceDockerConfigsConsoleOutputWide{}.Map(v).Sort()
	default:
		s := fmt.Sprintf("Not able to print to console! Unknow type %s.", v)
		return errors.New(s)
	}

	printer := tableprinter.New(c.Target)
	printer.HeaderLine = false

	printer.Print(out)
	return nil
}

func SortedString(m map[string]string) string {
	var s string

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		s += fmt.Sprintf("%s:%s\n", key, m[key])
	}

	return s[:len(s)-1]
}
