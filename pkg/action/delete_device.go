package action

import (
	"fmt"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/service"
)

func DeleteDevice(s *service.Service, ids []string) error {
	for _, d := range ids {
		msg, err := deleteDevice(s.Client, d)
		if err != nil {
			return err
		}
		fmt.Println(msg)
	}
	return nil
}

func deleteDevice(client *client.APIClient, deviceID string) (string, error) {
	_, err := client.DeleteRequest(api.DevicesEndpoint, deviceID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Device %s deleted successfully", deviceID), nil
}
