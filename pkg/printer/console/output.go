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
	default:
		s := fmt.Sprintf("Not able to print to console! Unknow type %s.", v)
		return errors.New(s)
	}

	printer := tableprinter.New(c.Target)
	printer.HeaderLine = false

	printer.Print(out)
	return nil
}
