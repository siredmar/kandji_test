package printer

import (
	"fmt"
	"strings"

	api "github.com/grid-x/gxctl/pkg/api"
)

type DeviceConfigMapConsoleOutput struct {
	ID         string `header:"ID"`
	Data       string `header:"Data"`
	BinaryData string `header:"Binary Data"`
}

type DeviceConfigMapConsoleOutputWide struct {
	ID         string `header:"ID"`
	Data       string `header:"Data"`
	Contents   string `header:"Contents"`
	BinaryData string `header:"Binary Data"`
}

func id(d api.DeviceConfigMap) string {
	if d.Spec.Immutable == nil || !(*d.Spec.Immutable) {
		return d.Metadata.ID
	}
	return "*" + d.Metadata.ID
}

func (o DeviceConfigMapConsoleOutput) Map(d api.DeviceConfigMap) DeviceConfigMapConsoleOutput {
	o.ID = id(d)

	if len(d.Spec.Data) > 0 {
		var b strings.Builder
		for k := range d.Spec.Data {
			fmt.Fprintf(&b, "%v\n", k)
		}
		o.Data = b.String()
	}

	if len(d.Spec.BinaryData) > 0 {
		var b strings.Builder
		for k := range d.Spec.BinaryData {
			fmt.Fprintf(&b, "%v\n", k)
		}
		o.BinaryData = b.String()
	}

	return o
}

func (o DeviceConfigMapConsoleOutputWide) Map(d api.DeviceConfigMap) []DeviceConfigMapConsoleOutputWide {
	var output []DeviceConfigMapConsoleOutputWide
	first := true
	i := 0
	bd := make([]string, len(d.Spec.BinaryData))
	for k := range d.Spec.BinaryData {
		bd[i] = k
		i++
	}

	j := 0
	for k, v := range d.Spec.Data {
		o := DeviceConfigMapConsoleOutputWide{}
		if first {
			o.ID = id(d)
			first = false
		}
		if j < i {
			o.BinaryData = bd[j]
			j++
		}
		o.Data = k
		o.Contents = truncate(v)

		output = append(output, o)
	}

	for j < i {
		o.BinaryData = bd[j]
		j++
		output = append(output, o)
	}

	return output
}

func truncate(in string) string {
	if len(in) < 50 {
		return in
	}

	return in[:50] + "..."
}
