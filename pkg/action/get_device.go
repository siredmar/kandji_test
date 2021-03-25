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
	showPublicIP bool,
	showPublicKey bool,
	deploymentID string,
	showAll bool,
	ids []string,
) error {
	printerConfig := print.PrintConfig{
		OutputFormat: outputType,
		SortBy:       sortBy,
		ShowAll:      showAll,
	}

	responseList := make(map[string]getDevicesResponse)
	var err error
	if deploymentID != "" {
		//Get devices by Deployment ID
		devices, err := getDevicesByDeploymentId(s.Client, deploymentID)
		if err != nil {
			return err
		}

		responseList["default"] = getDevicesResponse{device: devices}
	} else if len(ids) > 0 {
		resp := getDevicesResponse{}
		for _, id := range ids {
			device, err := getDeviceById(s.Client, id, nil)
			if err != nil {
				return err
			}

			resp.device.Devices = append(resp.device.Devices, device)
		}
		responseList["default"] = resp

		// If just looking for one specific device, make sure to print it even if it is offline
		printerConfig.ShowAll = true
	} else {
		//List all devices
		responseList, err = getDevices(s.Client, "", filter.NewCompositeFilter(filter.NewLabelFilter(label), filter.NewSerialnumberFilter(serial)))
		if err != nil {
			return err
		}
	}

	for _, resp := range responseList {
		if printerConfig.OutputFormat == print.Console || printerConfig.OutputFormat == print.ConsoleWide {
			if len(responseList) > 1 {
				if resp.err != nil {
					fmt.Printf("%v:\n%v\n\n", resp.profile, resp.err)
					continue
				} else {
					fmt.Println(resp.profile)
				}
			} else {
				if resp.err != nil {
					fmt.Println(resp.err)
				}
			}

			if showPublicIP {
				for _, device := range resp.device.Devices {
					if device.Status.Info != nil && device.Status.Info.PublicIP != nil {
						fmt.Println(*device.Status.Info.PublicIP)
					}
				}
				return nil
			}

			if showPublicKey {
				for _, device := range resp.device.Devices {
					if device.Status.Info != nil && device.Status.Info.PublicIP != nil {
						fmt.Println(*device.Spec.PublicKey)
					}
				}
				return nil
			}

			if err := s.Printer.Print(resp.device, printerConfig); err != nil {
				return err
			}
		}
	}
	if printerConfig.OutputFormat == print.JSON || printerConfig.OutputFormat == print.YAML {
		// raw
		var devices []api.Device
		for _, resp := range responseList {
			devices = append(devices, resp.device.Devices...)
		}

		if err := s.Printer.Print(devices, printerConfig); err != nil {
			return err
		}
	}

	return nil
}

type getDevicesResponse struct {
	device  api.Devices
	profile string
	err     error
}

func getDevices(c *client.APIClient, sn string, f filter.Filter) (map[string]getDevicesResponse, error) {
	var responses []client.RequestResult
	var err error

	if sn == "" {
		responses, err = c.GetMultiRequest(api.DevicesEndpoint)
	} else {
		responses, err = c.GetMultiRequest(fmt.Sprintf("%v?filter=serialnumber:%v", api.DevicesEndpoint, strings.ToUpper(sn)))
	}
	if err != nil {
		return nil, err
	}

	responseList := make(map[string]getDevicesResponse)
	for _, response := range responses {
		resp := getDevicesResponse{}
		resp.profile = response.Profile
		if response.Err != nil {
			resp.err = response.Err
		}

		deviceList, err := api.NewDevices(response.Body, false)
		if err != nil {
			resp.err = err
		}

		if f != nil {
			deviceList = filter.FilterDevices(deviceList, f)
		}

		if resp.err == nil && deviceList.IsEmpty() {
			resp.err = errors.E(
				errors.NotExists,
				"no devices found",
			)
		} else {
			resp.device = deviceList
		}

		responseList[response.Profile] = resp
	}

	return responseList, nil
}

func getDeviceById(client *client.APIClient, id string, deviceIds []string) (api.Device, error) {
	if len(id) != 36 && deviceIds == nil {
		responseList, err := getDevices(client, "", nil)
		if err != nil {
			return api.Device{}, err
		}
		for _, r := range responseList {
			deviceIds = append(deviceIds, r.device.GetIds()...)
		}
	}

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

func getDeviceBySN(client *client.APIClient, sn string) (map[string]getDevicesResponse, error) {
	return getDevices(client, sn, nil)
}

func getDevicesByDeploymentId(client *client.APIClient, deploymentID string) (api.Devices, error) {
	endpoint := fmt.Sprintf("%s/%s/devices", api.DeploymentsEndpoint, deploymentID)
	response, err := client.GetRequest(endpoint)
	if err != nil {
		return api.Devices{}, err
	}

	devices, err := api.NewDevices(response, false)
	if err != nil {
		return devices, err
	}

	return devices, nil
}
