package action

import (
	"fmt"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
	"github.com/grid-x/gxctl/pkg/service"
)

func Delete(s *service.Service, deleteCmdFileName string) error {
	contents, err := api.GetFilesContentsToProcess(deleteCmdFileName)
	if err != nil {
		return err
	}

	resources := make(map[string]api.Resource, len(contents))
	for _, c := range contents {
		res, resID, err := checkResourceFile(c, false)
		if err != nil {
			return err
		}
		resources[resID] = res
	}

	for resID, res := range resources {
		if err := deleteResource(resID, res, s.Client); err != nil {
			return err
		}
	}

	return nil
}

func deleteResource(resID string, res api.Resource, client *client.APIClient) error {
	fmt.Printf("deleting resource %s… ", resID)

	switch res.(type) {
	case *api.Application:
		if _, err := client.DeleteRequest(api.ApplicationsEndpoint, resID); err != nil {
			return err
		}
		break

	case *api.Device:
		if _, err := client.DeleteRequest(api.DevicesEndpoint, resID); err != nil {
			return err
		}
		break

	case *api.Deployment:
		if _, err := client.DeleteRequest(api.DeploymentsEndpoint, resID); err != nil {
			return err
		}
		break

	default:
		return errors.E(
			errors.NotImplemented,
			fmt.Sprintf("Unsupported type: %T", res),
		)
	}

	fmt.Print("success\n")

	return nil
}
