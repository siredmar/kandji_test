package action

import (
	"fmt"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
	print "github.com/grid-x/gxctl/pkg/printer"
	"github.com/grid-x/gxctl/pkg/service"
)

func ConfigGetDocker(s *service.Service, outputType string, ids []string) error {
	printerConfig := print.Printconfig{
		OutputFormat: outputType,
	}

	if len(ids) > 0 {
		//Get multiple dockerConfigs
		//Lookup all existing dockerConfigs to validate ids and autocomplete them if necessary
		dockerconfigs, err := getDockerConfigs(s.Client)
		if err != nil {
			return err
		}
		dockerConfigIDs := dockerconfigs.GetIds()

		for _, a := range ids {
			config, err := getDockerConfigById(s.Client, a, dockerConfigIDs)
			if err != nil {
				return err
			}

			if err := s.Printer.Print(config, printerConfig); err != nil {
				return err
			}
		}
	} else {
		//List all devices
		dockerConfigs, err := getDockerConfigs(s.Client)
		if err != nil {
			return err
		}

		if err := s.Printer.Print(dockerConfigs, printerConfig); err != nil {
			return err
		}
	}
	return nil
}

func getDockerConfigs(client *client.APIClient) (api.DockerConfigs, error) {
	response, err := client.GetRequest(api.DockerConfigsEndpoint)
	if err != nil {
		return api.DockerConfigs{}, err
	}

	dockerConfigList, err := api.NewDockerConfigs(response, false)
	if err != nil {
		return dockerConfigList, err
	}

	if dockerConfigList.IsEmpty() {
		return dockerConfigList, errors.E(
			errors.NotExists,
			"no dockerConfigs found",
		)
	}

	return dockerConfigList, nil
}

func getDockerConfigById(client *client.APIClient, id string, dockerConfigIDs []string) (api.DockerConfig, error) {
	configID, err := api.LookupID(id, dockerConfigIDs)
	if err != nil {
		return api.DockerConfig{}, err
	}

	endpoint := fmt.Sprintf("%s/%s", api.DockerConfigsEndpoint, configID)
	response, err := client.GetRequest(endpoint)
	if err != nil {
		return api.DockerConfig{}, err
	}

	config, err := api.NewDockerConfig(response, false)
	if err != nil {
		return config, err
	}

	return config, nil
}
