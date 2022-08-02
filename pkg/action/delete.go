package action

import (
	"fmt"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
	"github.com/grid-x/gxctl/pkg/service"
)

func Delete(s *service.Service, deleteCmdFileName string) error {
	resources, err := api.GetResources(deleteCmdFileName, false, false)
	if err != nil {
		return err
	}

	for _, r := range resources {
		message, err := deleteResource(r.ID, r.Res, s.Client)
		if err != nil {
			return err
		}
		fmt.Println(message)
	}

	return nil
}

func deleteResource(resID string, res api.Resource, client *client.APIClient) (string, error) {
	switch res.(type) {
	case *api.Application:
		if _, err := client.DeleteRequest(api.ApplicationsEndpoint, resID); err != nil {
			return "", err
		}
		return fmt.Sprintf("Application %s deleted successfully", resID), nil

	case *api.Device:
		if _, err := client.DeleteRequest(api.DevicesEndpoint, resID); err != nil {
			return "", err
		}
		return fmt.Sprintf("Device %s deleted successfully", resID), nil

	case *api.Deployment:
		if _, err := client.DeleteRequest(api.DeploymentsEndpoint, resID); err != nil {
			return "", err
		}
		return fmt.Sprintf("Deployment %s deleted successfully", resID), nil

	case *api.DeviceConfigMap:
		if _, err := client.DeleteRequest(api.DeviceConfigMapsEndpoint, resID); err != nil {
			return "", err
		}
		return fmt.Sprintf("DeviceConfigMap %s deleted successfully", resID), nil

	default:
		return "", errors.E(
			errors.NotImplemented,
			fmt.Sprintf("Unsupported type: %T", res),
		)
	}
}
