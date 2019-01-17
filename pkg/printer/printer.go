package printer

import (
	errors "github.com/grid-x/gxctl/pkg/error"
	console "github.com/grid-x/gxctl/pkg/printer/console"
	json "github.com/grid-x/gxctl/pkg/printer/json"
	"os"
)

const (
	Console string = ""
	JSON    string = "json"
	YAML    string = "yaml"
)

type Printer struct {
	Console *console.ConsolePrinter
	JSON    *json.JSONPrinter
}

func NewPrinter() *Printer {
	return &Printer{
		Console: console.NewConsolePrinter(os.Stdout),
		JSON:    json.NewJSONPrinter(os.Stdout),
	}
}

func (p *Printer) Print(d interface{}, outputFormat string) error {
	switch outputFormat {
	case JSON:
		return p.JSON.Print(d)
	case YAML:
		return errors.NotImplementedError("Getting yaml output")
	case Console:
		return p.Console.Print(d)
	default:
		return p.JSON.Print(d)
	}
}
