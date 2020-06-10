package action

import (
	"fmt"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
	print "github.com/grid-x/gxctl/pkg/printer"
	"github.com/grid-x/gxctl/pkg/service"
)

func GetDevice(s *service.Service, outputType string, sortBy string, showDockerConfig bool, showPods bool, showPublicKey bool, showAll bool, ids []string) error {
	printerConfig := print.Printconfig{
		OutputFormat: outputType,
		SortBy:       sortBy,
		ShowAll:      showAll,
	}

	if len(ids) > 0 {
		//Get multiple devices
		//Lookup all existing devices to validate ids and autocomplete them if necessary
		devices, err := getDevices(s.Client)
		if err != nil {
			return err
		}
		deviceIDs := devices.GetIds()

		for _, a := range ids {
			if showPublicKey {
				// Just show publickey
				device, err := getDeviceById(s.Client, a, deviceIDs)
				if err != nil {
					return err
				}
				fmt.Println(*device.Spec.PublicKey)
			} else if showDockerConfig {
				// Just show dockerconfig
				configs, err := getDeviceDockerConfigs(s.Client, a, deviceIDs)
				if err != nil {
					return err
				}

				if err := s.Printer.Print(configs, printerConfig); err != nil {
					return err
				}
			} else if showPods {
				// Just show pods
				pods, err := getDevicePods(s.Client, a, deviceIDs)
				if err != nil {
					return err
				}

				if err := s.Printer.Print(pods, printerConfig); err != nil {
					return err
				}
			} else {
				// Print device
				device, err := getDeviceById(s.Client, a, deviceIDs)
				if err != nil {
					return err
				}

				if err := s.Printer.Print(device, printerConfig); err != nil {
					return err
				}
			}
		}
	} else {
		//List all devices
		devices, err := getDevices(s.Client)
		if err != nil {
			return err
		}

		if err := s.Printer.Print(devices, printerConfig); err != nil {
			return err
		}
	}

	return nil

}

func getDevices(client *client.APIClient) (api.Devices, error) {
	response, err := client.GetRequest(api.DevicesEndpoint)
	if err != nil {
		return api.Devices{}, err
	}

	deviceList, err := api.NewDevices(response, false)
	if err != nil {
		return deviceList, err
	}

	if deviceList.IsEmpty() {
		return deviceList, errors.E(
			errors.NotExists,
			"no devices found",
		)
	}

	return deviceList, nil
}

func getDeviceById(client *client.APIClient, id string, deviceIds []string) (api.Device, error) {
	deviceID, err := api.LookupID(id, deviceIds)
	if err != nil {
		return api.Device{}, err
	}

	endpoint := fmt.Sprintf("%s/%s", api.DevicesEndpoint, deviceID)
	response, err := client.GetRequest(endpoint)
	if err != nil {
		return api.Device{}, err
	}

	device, err := api.NewDevice(response, false)
	if err != nil {
		return device, err
	}

	return device, nil
}

func getDeviceDockerConfigs(client *client.APIClient, id string, deviceIds []string) (api.DeviceDockerConfigs, error) {
	deviceID, err := api.LookupID(id, deviceIds)
	if err != nil {
		return api.DeviceDockerConfigs{}, err
	}

	endpoint := fmt.Sprintf("%s/%s/devicedockerconfigs", api.DevicesEndpoint, deviceID)
	response, err := client.GetRequest(endpoint)
	if err != nil {
		return api.DeviceDockerConfigs{}, err
	}

	configs, err := api.NewDeviceDockerConfigs(response, false)
	if err != nil {
		return configs, err
	}

	return configs, nil
}

func getDevicePods(client *client.APIClient, id string, deviceIds []string) (api.Pods, error) {
	deviceID, err := api.LookupID(id, deviceIds)
	if err != nil {
		return api.Pods{}, err
	}

	endpoint := fmt.Sprintf("%s/%s/pods", api.DevicesEndpoint, deviceID)
	response, err := client.GetRequest(endpoint)
	if err != nil {
		return api.Pods{}, err
	}

	pods, err := api.NewPods(response, false)
	if err != nil {
		return pods, err
	}

	return pods, nil
}
