package action

import (
	"fmt"

	maintenanceApi "github.com/grid-x/ds-api-types/management/2019-11-04/maintenance"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/service"
)

func Restart(s *service.Service, id string) error {
	d, err := getDeviceById(s.Client, id, nil)
	if err != nil {
		return err
	}

	message, err := restartDevice(s.Client, d.Metadata.ID)
	if err != nil {
		return err
	}

	fmt.Println(message)
	return nil
}

func restartDevice(client *client.APIClient, devID string) (string, error) {
	d := api.CreateMaintenanceTask{
		Spec: &maintenanceApi.MaintenanceTaskSpec{
			Type:     maintenanceApi.MaintenanceTaskTypeRestart,
			DeviceID: devID,
		},
	}

	message, err := createResource(devID, d, client)
	if err != nil {
		return "", err
	}

	return message, nil
}
