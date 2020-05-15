package action

import (
	"fmt"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/service"
)

func CreateApplication(s *service.Service, createApplicationCmdName string) error {
	d := api.Application{}

	if createApplicationCmdName != "" {
		d.Name = createApplicationCmdName
	}

	message, err := createResource(createApplicationCmdName, d, s.Client)
	if err != nil {
		return err
	}

	fmt.Println(message)
	return nil
}
