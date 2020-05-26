package action

import (
	"fmt"

	maintenanceApi "github.com/grid-x/ds-api-types/management/2019-11-04/maintenance"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/service"
)

func Restart(s *service.Service, id string) error {
	//Lookup all existing devices to validate ids and autocomplete them if necessary
	devices, err := getDevices(s.Client)
	if err != nil {
		return err
	}
	deviceIDs := devices.GetIds()

	message, err := restartDevice(s.Client, id, deviceIDs)
	if err != nil {
		return err
	}

	fmt.Println(message)
	return nil
}

func restartDevice(client *client.APIClient, id string, ids []string) (string, error) {
	devID := id
	if ids != nil {
		var err error
		devID, err = api.LookupID(id, ids)
		if err != nil {
			return "", err
		}
	}

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
