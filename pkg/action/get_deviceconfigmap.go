package action

import (
	"fmt"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/printer"
	"github.com/grid-x/gxctl/pkg/service"
)

func GetDeviceConfigMap(s *service.Service, outputType string, sortBy string, ids []string) error {
	printerConfig := printer.PrintConfig{
		OutputFormat: outputType,
		SortBy:       sortBy,
	}

	if len(ids) == 1 {
		dcm, err := getDeviceConfigMapByID(s.Client, ids[0])
		if err != nil {
			return err
		}
		return s.Printer.Print(dcm, printerConfig)
	}

	dcms, err := getDeviceConfigMaps(s.Client)
	if err != nil {
		return err
	}

	if printerConfig.OutputFormat == printer.JSON || printerConfig.OutputFormat == printer.YAML {
		return s.Printer.Print(dcms.DeviceConfigMaps, printerConfig)
	}

	return s.Printer.Print(dcms, printerConfig)
}

func getDeviceConfigMapByID(client *client.APIClient, id string) (api.DeviceConfigMap, error) {
	response, err := client.GetRequest(fmt.Sprintf("%s/%s", api.DeviceConfigMapsEndpoint, id))
	if err != nil {
		return api.DeviceConfigMap{}, err
	}

	dcm, err := api.NewDeviceConfigMap(response, false)
	if err != nil {
		return api.DeviceConfigMap{}, err
	}

	return dcm, nil
}

func getDeviceConfigMaps(client *client.APIClient) (api.DeviceConfigMaps, error) {
	response, err := client.GetRequest(api.DeviceConfigMapsEndpoint)
	if err != nil {
		return api.DeviceConfigMaps{}, err
	}

	dcms, err := api.NewDeviceConfigMaps(response, false)
	if err != nil {
		return api.DeviceConfigMaps{}, err
	}

	return dcms, nil
}
