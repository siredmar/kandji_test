package printer

import (
	"errors"
	"fmt"
	"github.com/landoop/tableprinter"
	"io"

	api "github.com/grid-x/gxctl/pkg/api"
)

type ConsolePrinter struct {
	Target io.Writer
}

func NewConsolePrinter(t io.Writer) *ConsolePrinter {
	return &ConsolePrinter{
		Target: t,
	}
}

type printClient interface {
	Print(v interface{})
}

func (c *ConsolePrinter) Print(v interface{}) error {
	var out interface{}

	switch v := v.(type) {
	case api.Device:
		out = DeviceConsoleOutput{}.Map(v)
	case api.Devices:
		out = DevicesConsoleOutput{}.Map(v).Sort()
	case api.Pod:
		out = PodConsoleOutput{}.Map(v)
	case api.Pods:
		out = PodsConsoleOutput{}.Map(v).Sort()
	case api.Deployment:
		out = DeploymentConsoleOutput{}.Map(v)
	case api.Deployments:
		out = DeploymentsConsoleOutput{}.Map(v).Sort()
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

func (c *ConsolePrinter) PrintWide(v interface{}) error {
	var out interface{}

	switch v := v.(type) {
	case api.Device:
		out = DeviceConsoleOutputWide{}.Map(v)
	case api.Devices:
		out = DevicesConsoleOutputWide{}.Map(v).Sort()
	case api.Pod:
		out = PodConsoleOutputWide{}.Map(v)
	case api.Pods:
		out = PodsConsoleOutputWide{}.Map(v).Sort()
	case api.Deployment:
		out = DeploymentConsoleOutput{}.Map(v) //TODO make wide mapping
	case api.Deployments:
		out = DeploymentsConsoleOutput{}.Map(v).Sort() //TODO make wide mapping
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
