package action

import (
	"fmt"
	"strings"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
	"github.com/grid-x/gxctl/pkg/filter"
	print "github.com/grid-x/gxctl/pkg/printer"
	"github.com/grid-x/gxctl/pkg/service"
)

func GetDevice(
	s *service.Service,
	outputType string,
	label string,
	serial string,
	sortBy string,
	showDeployments bool,
	showDockerConfig bool,
	showPods bool,
	showPublicIP bool,
	showPublicKey bool,
	showAll bool,
	ids []string,
) error {
	printerConfig := print.PrintConfig{
		OutputFormat: outputType,
		SortBy:       sortBy,
		ShowAll:      showAll,
		//Filter:       filter.NewCompositeFilter(filter.NewLabelFilter(label), filter.NewSerialnumberFilter(serial)),
	}

	//List all devices
	var devices api.Devices
	rawDevices, err, errs := getDevices(s.Client)
	if err != nil {
		return err
	}
	rawDeviceIDs := rawDevices.GetIds()

	if printerConfig.OutputFormat == print.Console || printerConfig.OutputFormat == print.ConsoleWide {
		for _, err := range errs {
			if err.err != nil {
				fmt.Printf("%v: %v\n", err.profile, err.err)
			}
		}
	}

	if len(ids) > 0 {
		for _, id := range ids {
			device, err := getDeviceById(s.Client, id, rawDeviceIDs)
			if err != nil {
				return err
			}

			devices.Devices = append(devices.Devices, device)
		}
	} else {
		devices = rawDevices
	}

	devices = filter.FilterDevices(devices, filter.NewCompositeFilter(filter.NewLabelFilter(label), filter.NewSerialnumberFilter(serial)))

	if showPublicIP {
		for _, device := range devices.Devices {
			if device.Status.Info != nil && device.Status.Info.PublicIP != nil {
				fmt.Println(*device.Status.Info.PublicIP)
			}
		}
		return nil
	}

	if showPublicKey {
		for _, device := range devices.Devices {
			if device.Status.Info != nil && device.Status.Info.PublicIP != nil {
				fmt.Println(*device.Spec.PublicKey)
			}
		}
		return nil
	}

	if showDeployments {
		if len(devices.Devices) > 1 {
			return errors.E(
				errors.Invalid,
				"Too many devices - Deployments can just be shown for a single device",
				nil,
			)
		}

		// Just show deployments
		pods, err := getDevicePods(s.Client, devices.Devices[0].Metadata.ID)
		if err != nil {
			return err
		}

		var deploymentIDs []string

		for _, pod := range pods.Pods {

			if pod.Metadata.Annotations == nil {
				continue
			}

			id, ok := pod.Metadata.Annotations["gridx.ai/deployment"]

			if ok {
				deploymentIDs = append(deploymentIDs, id)
			}
		}

		var deployments []api.Deployment

		for _, id := range deploymentIDs {
			deployment, err := getDeploymentById(s.Client, id, deploymentIDs)
			if err != nil {
				return err
			}

			deployments = append(deployments, deployment)
		}

		if err := s.Printer.Print(api.Deployments{Deployments: deployments}, printerConfig); err != nil {
			return err
		}
	}

	if showDockerConfig {
		if len(devices.Devices) > 1 {
			return errors.E(
				errors.Invalid,
				"Too many devices - Docker-Configs can just be shown for a single device",
				nil,
			)
		}

		// Just show dockerconfig
		configs, err := getDeviceDockerConfigs(s.Client, devices.Devices[0].Metadata.ID)
		if err != nil {
			return err
		}

		if err := s.Printer.Print(configs, printerConfig); err != nil {
			return err
		}

		return nil
	}

	if showPods {
		if len(devices.Devices) > 1 {
			return errors.E(
				errors.Invalid,
				"Too many devices - Pods can just be shown for a single device",
				nil,
			)
		}

		// Just show pods
		pods, err := getDevicePods(s.Client, devices.Devices[0].Metadata.ID)
		if err != nil {
			return err
		}

		if err := s.Printer.Print(pods, printerConfig); err != nil {
			return err
		}

		return nil
	}

	if err := s.Printer.Print(devices, printerConfig); err != nil {
		return err
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
		if response.Err != nil {
			errs[i].err = response.Err
			continue
		}

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

func getDeviceDockerConfigs(client *client.APIClient, deviceID string) (api.DeviceDockerConfigs, error) {
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

func getDevicePods(client *client.APIClient, deviceID string) (api.Pods, error) {
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
