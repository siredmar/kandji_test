package printer

import (
	"os"

	console "github.com/grid-x/gxctl/pkg/printer/console"
	json "github.com/grid-x/gxctl/pkg/printer/json"
	yaml "github.com/grid-x/gxctl/pkg/printer/yaml"
)

const (
	Console     string = ""
	ConsoleWide string = "wide"
	JSON        string = "json"
	YAML        string = "yaml"
)

type Printer struct {
	Console *console.ConsolePrinter
	JSON    *json.JSONPrinter
	YAML    *yaml.YAMLPrinter
}

func NewPrinter() *Printer {
	return &Printer{
		Console: console.NewConsolePrinter(os.Stdout),
		JSON:    json.NewJSONPrinter(os.Stdout),
		YAML:    yaml.NewYAMLPrinter(os.Stdout),
	}
}

func (p *Printer) Print(d interface{}, outputFormat string) error {
	switch outputFormat {
	case JSON:
		return p.JSON.Print(d)
	case YAML:
		return p.YAML.Print(d)
	case Console:
		return p.Console.Print(d)
	case ConsoleWide:
		return p.Console.PrintWide(d)
	default:
		return p.JSON.Print(d)
	}
}
