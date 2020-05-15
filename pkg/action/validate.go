package action

import (
	"fmt"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/service"
)

func Validate(s *service.Service, fileName string) error {
	contents, err := api.GetFilesContentsToProcess(fileName)
	if err != nil {
		return err
	}

	for n, c := range contents {
		if err := validate(n, c, s.Client); err != nil {
			return err
		}
	}

	return nil
}

func validate(filename string, content []byte, client *client.APIClient) error {
	fmt.Printf("Validating file %s\n", filename)

	_, _, err := checkResourceFile(content, true)
	if err != nil {
		return err
	}

	fmt.Println("Valid Resource")
	return nil
}
