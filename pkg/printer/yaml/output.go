package printer

import (
	"io"

	"github.com/ghodss/yaml"
)

type YAMLPrinter struct {
	Target io.Writer
}

func NewYAMLPrinter(t io.Writer) *YAMLPrinter {
	return &YAMLPrinter{
		Target: t,
	}
}

func (p *YAMLPrinter) Print(v interface{}) error {
	out, err := yaml.Marshal(v)

	if err != nil {
		return err
	}

	_, err = p.Target.Write(out)

	if err != nil {
		return err
	}
	return nil
}
