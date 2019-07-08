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

type Printconfig struct {
	OutputFormat string
	SortBy       string
	ShowAll      bool
}

func NewPrinter() *Printer {
	return &Printer{
		Console: console.NewConsolePrinter(os.Stdout),
		JSON:    json.NewJSONPrinter(os.Stdout),
		YAML:    yaml.NewYAMLPrinter(os.Stdout),
	}
}

func (p *Printer) Print(d interface{}, config Printconfig) error {
	switch config.OutputFormat {
	case JSON:
		return p.JSON.Print(d)
	case YAML:
		return p.YAML.Print(d)
	case Console:
		c := console.ConsolePrintconfig{
			SortBy:  config.SortBy,
			ShowAll: config.ShowAll,
		}
		return p.Console.Print(d, c)
	case ConsoleWide:
		c := console.ConsolePrintconfig{
			SortBy:  config.SortBy,
			ShowAll: config.ShowAll,
		}
		return p.Console.PrintWide(d, c)
	default:
		return p.JSON.Print(d)
	}
}
