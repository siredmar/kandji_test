package printer

import (
	"encoding/json"
	"io"
)

type JSONPrinter struct {
	Target io.Writer
}

func NewJSONPrinter(t io.Writer) *JSONPrinter {
	return &JSONPrinter{
		Target: t,
	}
}

func (p *JSONPrinter) Print(v interface{}) error {
	out, err := json.MarshalIndent(v, "", "    ")

	if err != nil {
		return err
	}

	_, err = p.Target.Write(out)

	if err != nil {
		return err
	}
	return nil
}
