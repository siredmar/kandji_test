package action

import (
	"fmt"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/service"
)

func Validate(s *service.Service, fileName string) error {
	fmt.Printf("Validating file %s\n", fileName)

	_, err := api.GetResources(fileName, true, true)
	if err != nil {
		return err
	}

	fmt.Println("Valid Resource")
	return nil
}
