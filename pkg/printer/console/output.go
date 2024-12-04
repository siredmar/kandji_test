package printer

import (
	"errors"
	"fmt"
	"io"
	"sort"

	"github.com/landoop/tableprinter"

	api "github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/filter"
)

type ConsolePrinter struct {
	Target io.Writer
}

type ConsolePrintConfig struct {
	Filter filter.Filter
	SortBy string
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
		out = DevicesConsoleOutput{}.Inject(v).Filter(config.Filter).Sort(config.SortBy).Map()
	case api.Pod:
		out = PodConsoleOutput{}.Map(v)
	case api.DeviceConfigMap:
		out = DeviceConfigMapConsoleOutput{}.Map(v)
	case api.DeviceConfigMaps:
		out = DeviceConfigMapsConsoleOutput{}.Inject(v).Map()
	case api.Pods:
		out = PodsConsoleOutput{}.Inject(v).Sort(config.SortBy).Map()
	case api.Deployment:
		out = DeploymentConsoleOutput{}.Map(v)
	case api.Deployments:
		out = DeploymentsConsoleOutput{}.Inject(v).Sort(config.SortBy).Map()
	case api.Application:
		out = ApplicationConsoleOutput{}.Map(v)
	case api.Applications:
		out = ApplicationsConsoleOutput{}.Map(v).Sort()
	case api.MaintenanceTask:
		out = MaintenanceConsoleOutput{}.Map(v)
	case api.MaintenanceTasks:
		out = MaintenancesConsoleOutput{}.Map(v).Sort()
	default:
		s := fmt.Sprintf("Not able to print to console! Unknown type %T.", v)
		return errors.New(s)
	}

	printer := tableprinter.New(c.Target)
	printer.HeaderLine = false

	printer.Print(out)
	return nil
}

func noRowLengthTitle(_ int) bool { return false }

func (c *ConsolePrinter) PrintWide(v interface{}, config ConsolePrintConfig) error {
	printer := tableprinter.New(c.Target)
	printer.HeaderLine = false

	var out interface{}

	switch v := v.(type) {
	case api.Device:
		out = DeviceConsoleOutputWide{}.Map(v)
	case api.Devices:
		out = DevicesConsoleOutputWide{}.Inject(v).Filter(config.Filter).Sort(config.SortBy).Map()
	case api.DeviceConfigMap:
		printer.RowLengthTitle = noRowLengthTitle
		out = DeviceConfigMapConsoleOutputWide{}.Map(v)
	case api.DeviceConfigMaps:
		printer.RowLengthTitle = noRowLengthTitle
		out = DeviceConfigMapsConsoleOutputWide{}.Inject(v).Map()
	case api.Pod:
		out = PodConsoleOutputWide{}.Map(v)
	case api.Pods:
		out = PodsConsoleOutputWide{}.Inject(v).Sort(config.SortBy).Map()
	case api.Deployment:
		out = DeploymentConsoleOutputWide{}.Map(v)
	case api.Deployments:
		out = DeploymentsConsoleOutputWide{}.Inject(v).Sort(config.SortBy).Map()
	case api.Application:
		out = ApplicationConsoleOutput{}.Map(v) // TODO make wide mapping
	case api.Applications:
		out = ApplicationsConsoleOutput{}.Map(v).Sort() // TODO make wide mapping
	case api.MaintenanceTask:
		out = MaintenanceConsoleOutputWide{}.Map(v)
	case api.MaintenanceTasks:
		out = MaintenancesConsoleOutputWide{}.Map(v).Sort()
	case api.DeviceLogs:
		out = DeviceLogsConsoleOutputWide{}.Map(v)
	case api.DevicesLogs:
		out = DevicesLogsConsoleOutputWide{}.Map(v)
	default:
		s := fmt.Sprintf("Not able to print to console! Unknown type %T.", v)
		return errors.New(s)
	}

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
