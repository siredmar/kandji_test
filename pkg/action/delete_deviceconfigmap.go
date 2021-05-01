package action

import (
	"fmt"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/service"
)

func DeleteDeviceConfigMap(s *service.Service, ids []string) error {
	for _, d := range ids {
		msg, err := deleteDeviceConfigMap(s.Client, d)
		if err != nil {
			return err
		}
		fmt.Println(msg)
	}
	return nil
}

func deleteDeviceConfigMap(client *client.APIClient, dcmID string) (string, error) {
	_, err := client.DeleteRequest(api.DeviceConfigMapsEndpoint, dcmID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("DeviceConfigMap %s deleted successfully", dcmID), nil
}
