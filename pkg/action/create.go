package action

import (
	"fmt"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
	"github.com/grid-x/gxctl/pkg/service"
)

func Create(s *service.Service, fileName string) error {
	contents, err := api.GetFilesContentsToProcess(fileName)
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

	resourcesSorted := sortByKind(resources)

	for _, r := range resourcesSorted {
		if err := create(r.ID, r.Res, s.Client); err != nil {
			return err
		}
	}

	return nil
}

func create(resID string, res interface{}, client *client.APIClient) error {
	message, err := createResource(resID, res, client)
	if err != nil {
		return err
	}

	fmt.Println(message)
	return nil
}

func createResource(resID string, res interface{}, client *client.APIClient) (string, error) {
	switch res := res.(type) {
	case api.Device:
		response, err := client.PostRequest(api.DevicesEndpoint, res)
		if err != nil {
			return "", err
		}
		if _, err := api.NewDevice(response, false); err != nil {
			return "", err
		}
		return fmt.Sprintf("Device %s created successfully", resID), nil

	case api.Deployment:
		response, err := client.PostRequest(api.DeploymentsEndpoint, res)
		if err != nil {
			return "", err
		}
		if _, err := api.NewDeployment(response, false); err != nil {
			return "", err
		}
		return fmt.Sprintf("Deployment %s created successfully", resID), nil

	case api.Application:
		response, err := client.PostRequest(api.ApplicationsEndpoint, res)
		if err != nil {
			return "", err
		}
		if _, err := api.NewApplication(response, false); err != nil {
			return "", err
		}
		return fmt.Sprintf("Application %s created successfully", resID), nil

	case api.MaintenanceTask:
		response, err := client.PostRequest(api.MaintenanceEndpoint, res)
		if err != nil {
			return "", err
		}
		if _, err := api.NewMaintenanceTask(response, false); err != nil {
			return "", err
		}
		return fmt.Sprintf("Maintenance task %s created successfully", resID), nil

	case api.DockerConfig:
		response, err := client.PostRequest(api.DockerConfigsEndpoint, res)
		if err != nil {
			return "", err
		}
		if _, err := api.NewDockerConfig(response, false); err != nil {
			return "", err
		}
		return fmt.Sprintf("Docker config %s created successfully", resID), nil

	case api.CleanupConfig:
		response, err := client.PostRequest(api.CleanupConfigsEndpoint, res)
		if err != nil {
			return "", err
		}
		if _, err := api.NewCleanupConfig(response, false); err != nil {
			return "", err
		}
		return fmt.Sprintf("Cleanup config %s created successfully", resID), nil

	default:
		return "", errors.E(
			errors.NotImplemented,
			"Unsupported type",
		)
	}
}
