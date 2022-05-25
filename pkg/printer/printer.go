package printer

import (
	"os"

	"github.com/grid-x/gxctl/pkg/filter"
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

type PrintConfig struct {
	OutputFormat string
	Filter       filter.Filter
	SortBy       string
}

func NewPrinter() *Printer {
	return &Printer{
		Console: console.NewConsolePrinter(os.Stdout),
		JSON:    json.NewJSONPrinter(os.Stdout),
		YAML:    yaml.NewYAMLPrinter(os.Stdout),
	}
}

func (p *Printer) Print(d interface{}, config PrintConfig) error {
	switch config.OutputFormat {
	case JSON:
		return p.JSON.Print(d)
	case YAML:
		return p.YAML.Print(d)
	case Console:
		c := console.ConsolePrintConfig{
			Filter: config.Filter,
			SortBy: config.SortBy,
		}
		return p.Console.Print(d, c)
	case ConsoleWide:
		c := console.ConsolePrintConfig{
			Filter: config.Filter,
			SortBy: config.SortBy,
		}
		return p.Console.PrintWide(d, c)
	default:
		return p.JSON.Print(d)
	}
}
