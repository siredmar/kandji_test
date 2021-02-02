package action

import (
	"fmt"

	types "github.com/grid-x/ds-api-types"

	"github.com/grid-x/gxctl/pkg/api"

	"github.com/grid-x/gxctl/pkg/service"
)

func UpdateDevice(s *service.Service,
	id string,
	updateDeviceCmdMaintenanceWindow string,
	updateDeviceCmdMacAddress string,
	updateDeviceCmdLabels string,
	updateDeviceCmdAnnotations string,
) error {
	d, err := getDeviceById(s.Client, id, nil)
	if err != nil {
		return err
	}

	if updateDeviceCmdMaintenanceWindow != "" {
		w, err := types.NewMaintenanceWindow(updateDeviceCmdMaintenanceWindow)
		if err != nil {
			return err
		}
		d.Spec.MaintenanceWindow = w
	}
	if updateDeviceCmdMacAddress != "" {
		d.Spec.MACAddress = &updateDeviceCmdMacAddress
	}

	if updateDeviceCmdLabels != "" {
		m, err := api.ParseMetadataMap(updateDeviceCmdLabels)
		if err != nil {
			return err
		}

		d.Metadata.Labels = m
	}

	if updateDeviceCmdAnnotations != "" {
		a, err := api.ParseMetadataMap(updateDeviceCmdAnnotations)
		if err != nil {
			return err
		}

		d.Metadata.Annotations = a
	}

	message, err := updateResource(s.Client, d, d.Metadata.ID, nil)
	if err != nil {
		return err
	}

	fmt.Println(message)
	return nil
}
