package action

import (
	"fmt"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
	print "github.com/grid-x/gxctl/pkg/printer"
	"github.com/grid-x/gxctl/pkg/service"
)

func ConfigGetCleanup(s *service.Service, outputType string, ids []string) error {
	printerConfig := print.Printconfig{
		OutputFormat: outputType,
	}

	if len(ids) > 0 {
		//Get multiple cleanupConfigs
		//Lookup all existing cleanupConfigs to validate ids and autocomplete them if necessary
		cleanupconfigs, err := getCleanupConfigs(s.Client)
		if err != nil {
			return err
		}
		cleanupConfigIDs := cleanupconfigs.GetIds()

		for _, a := range ids {
			config, err := getCleanupConfigById(s.Client, a, cleanupConfigIDs)
			if err != nil {
				return err
			}

			if err := s.Printer.Print(config, printerConfig); err != nil {
				return err
			}
		}
	} else {
		//List all devices
		cleanupConfigs, err := getCleanupConfigs(s.Client)
		if err != nil {
			return err
		}

		if err := s.Printer.Print(cleanupConfigs, printerConfig); err != nil {
			return err
		}
	}
	return nil
}

func getCleanupConfigs(client *client.APIClient) (api.CleanupConfigs, error) {
	response, err := client.GetRequest(api.CleanupConfigsEndpoint)
	if err != nil {
		return api.CleanupConfigs{}, err
	}

	cleanupConfigList, err := api.NewCleanupConfigs(response)
	if err != nil {
		return cleanupConfigList, err
	}

	if cleanupConfigList.IsEmpty() {
		return cleanupConfigList, errors.E(
			errors.NotExists,
			"no cleanupConfigs found",
		)
	}

	return cleanupConfigList, nil
}

func getCleanupConfigById(client *client.APIClient, id string, cleanupConfigIDs []string) (api.CleanupConfig, error) {
	configID, err := api.LookupID(id, cleanupConfigIDs)
	if err != nil {
		return api.CleanupConfig{}, err
	}

	endpoint := fmt.Sprintf("%s/%s", api.CleanupConfigsEndpoint, configID)
	response, err := client.GetRequest(endpoint)
	if err != nil {
		return api.CleanupConfig{}, err
	}

	config, err := api.NewCleanupConfig(response)
	if err != nil {
		return config, err
	}

	return config, nil
}
