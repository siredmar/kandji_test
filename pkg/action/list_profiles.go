package action

import (
	"fmt"

	"github.com/grid-x/gxctl/pkg/printer"
	"github.com/grid-x/gxctl/pkg/service"
)

type profiles struct {
	Profiles []string `json:"profiles"`
}

func ListProfiles(s *service.Service, outputType string) error {
	printerConfig := printer.PrintConfig{
		OutputFormat: outputType,
	}

	profiles := profiles{
		Profiles: []string{},
	}
	for _, p := range s.Client.Auth.Profiles {
		profiles.Profiles = append(profiles.Profiles, p.Name)
	}

	switch printerConfig.OutputFormat {
	case printer.JSON, printer.YAML:
		return s.Printer.Print(profiles, printerConfig)

	default:
		for _, p := range profiles.Profiles {
			fmt.Println(p)
		}
	}
	return nil
}
