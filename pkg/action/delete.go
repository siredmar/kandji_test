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

	resources := make(map[string]interface{}, len(contents))
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

func deleteResource(resID string, res interface{}, client *client.APIClient) error {
	fmt.Printf("deleting resource %s… ", resID)

	switch res.(type) {
	case api.Application:
		if _, err := client.DeleteRequest(api.ApplicationsEndpoint, resID); err != nil {
			return err
		}
		break

	case api.Device:
		if _, err := client.DeleteRequest(api.DevicesEndpoint, resID); err != nil {
			return err
		}
		break

	case api.Deployment:
		if _, err := client.DeleteRequest(api.DockerConfigsEndpoint, resID); err != nil {
			return err
		}
		break

	case api.DockerConfig:
		if _, err := client.DeleteRequest(api.DockerConfigsEndpoint, resID); err != nil {
			return err
		}
		break

	case api.CleanupConfig:
		if _, err := client.DeleteRequest(api.CleanupConfigsEndpoint, resID); err != nil {
			return err
		}
		break

	default:
		return errors.E(
			errors.NotImplemented,
			"Unsupported type",
		)
	}

	fmt.Print("success\n")

	return nil
}
