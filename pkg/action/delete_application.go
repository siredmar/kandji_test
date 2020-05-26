package action

import (
	"fmt"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/service"
)

func DeleteApplication(s *service.Service, ids []string) error {
	for _, d := range ids {
		msg, err := deleteApp(s.Client, d)
		if err != nil {
			return err
		}
		fmt.Println(msg)
	}
	return nil
}

func deleteApp(client *client.APIClient, appID string) (string, error) {
	_, err := client.DeleteRequest(api.ApplicationsEndpoint, appID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("App %s deleted successfully", appID), nil
}
