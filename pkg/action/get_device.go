package action

import (
	"fmt"
	"strings"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
	print "github.com/grid-x/gxctl/pkg/printer"
	"github.com/grid-x/gxctl/pkg/service"
)

func GetDevice(
	s *service.Service,
	outputType string,
	sortBy string,
	showDockerConfig bool,
	showPods bool,
	showPublicIP bool,
	showPublicKey bool,
	showAll bool,
	ids []string,
) error {
	printerConfig := print.Printconfig{
		OutputFormat: outputType,
		SortBy:       sortBy,
		ShowAll:      showAll,
	}

	if len(ids) > 0 {
		//Get multiple devices
		//Lookup all existing devices to validate ids and autocomplete them if necessary
		devices, err, _ := getDevices(s.Client)
		if err != nil {
			return err
		}
		deviceIDs := devices.GetIds()

		for _, a := range ids {
			if showPublicIP {
				// Just show public IP
				device, err := getDeviceById(s.Client, a, deviceIDs)
				if err != nil {
					return err
				}
				if device.Status.Info != nil && device.Status.Info.PublicIP != nil {
					fmt.Println(*device.Status.Info.PublicIP)
				}
			} else if showPublicKey {
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
		devices, err, errs := getDevices(s.Client)
		if err != nil {
			return err
		}

		if printerConfig.OutputFormat == print.Console || printerConfig.OutputFormat == print.ConsoleWide {
			for _, err := range errs {
				if err.err != nil {
					fmt.Printf("%v: %v\n", err.profile, err.err)
				}
			}
		}

		if err := s.Printer.Print(devices, printerConfig); err != nil {
			return err
		}
	}

	return nil

}

type getDevicesError struct {
	profile string
	err     error
}

func getDevices(client *client.APIClient) (api.Devices, error, []getDevicesError) {
	devices := api.Devices{}

	responses, err := client.GetMultiRequest(api.DevicesEndpoint)
	if err != nil {
		return devices, err, nil
	}

	errs := make([]getDevicesError, len(responses))

	for i, response := range responses {
		errs[i].profile = response.Profile
		deviceList, err := api.NewDevices(response.Body, false)
		if err != nil {
			errs[i].err = err
			continue
		}

		if deviceList.IsEmpty() {
			errs[i].err = errors.E(
				errors.NotExists,
				"no devices found",
			)
		} else {
			for _, d := range deviceList.Devices {
				devices.Devices = append(devices.Devices, d)
			}
		}
	}

	return devices, nil, errs

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

func getDeviceBySN(client *client.APIClient, sn string) (api.Device, error, string, []getDevicesError) {
	responses, err := client.GetMultiRequest(fmt.Sprintf("%v?filter=serialnumber:%v", api.DevicesEndpoint, strings.ToUpper(sn)))
	if err != nil {
		return api.Device{}, err, "", nil
	}

	var devicesList []api.Device
	var profile string
	errs := make([]getDevicesError, len(responses))

	for i, r := range responses {
		errs[i] = getDevicesError{
			profile: r.Profile,
			err:     r.Err,
		}
		if r.Err == nil {
			devices, err := api.NewDevices(r.Body, false)
			if err != nil {
				errs[i].err = err
				continue
			}
			if len(devices.Devices) > 0 {
				profile = r.Profile
			}
			for _, d := range devices.Devices {
				devicesList = append(devicesList, d)
			}
		}
	}

	switch len(devicesList) {
	case 0:
		return api.Device{},
			errors.E(
				errors.NotExists,
				"no device found",
			), "", errs
	case 1:
		return devicesList[0], nil, profile, errs
	default:
		return api.Device{}, errors.E(errors.Invalid, "more than one device found"), "", errs
	}
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
