package action

import (
	"fmt"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/service"
)

func ConfigDeleteDocker(s *service.Service, ids []string) error {
	dockerConfigs, err := getDockerConfigs(s.Client)
	if err != nil {
		return err
	}
	dockerConfigIDs := dockerConfigs.GetIds()

	for _, d := range ids {
		msg, err := deleteDockerConfig(s.Client, d, dockerConfigIDs)
		if err != nil {
			return err
		}
		fmt.Println(msg)
	}
	return nil
}

func deleteDockerConfig(client *client.APIClient, configID string, dockerConfigIDs []string) (string, error) {
	configID, err := api.LookupID(configID, dockerConfigIDs)
	if err != nil {
		return "", err
	}

	_, err = client.DeleteRequest(api.DockerConfigsEndpoint, configID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Docker config %s deleted successfully", configID), nil
}
