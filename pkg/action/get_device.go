package action

import (
	"fmt"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
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
	ids []string,
) error {
	printerConfig := print.PrintConfig{
		OutputFormat: outputType,
		SortBy:       sortBy,
	}

	responseList := make(map[string]client.GetDevicesResponse)
	var err error
	if deploymentID != "" {
		//Get devices by Deployment ID
		devices, err := getDevicesByDeploymentId(s.Client, deploymentID)
		if err != nil {
			return err
		}

		responseList["default"] = client.GetDevicesResponse{Device: devices}
	} else if len(ids) > 0 {
		resp := client.GetDevicesResponse{}
		for _, id := range ids {
			device, err := client.GetDeviceById(s.Client, id, nil)
			if err != nil {
				return err
			}

			resp.Device.Devices = append(resp.Device.Devices, device)
		}
		responseList["default"] = resp
	} else if serial != "" {
		responseList, err = getDeviceBySN(s.Client, serial)
		if err != nil {
			return err
		}
	} else {
		//List all devices
		responseList, err = client.GetDevices(s.Client, "", filter.NewLabelFilter(label))
		if err != nil {
			return err
		}
	}

	for _, resp := range responseList {
		if printerConfig.OutputFormat == print.Console || printerConfig.OutputFormat == print.ConsoleWide {
			if len(responseList) > 1 {
				if resp.Err != nil {
					fmt.Printf("%v:\n%v\n\n", resp.Profile, resp.Err)
					continue
				} else {
					fmt.Println(resp.Profile)
				}
			} else {
				if resp.Err != nil {
					fmt.Println(resp.Err)
				}
			}

			if showPublicIP {
				for _, device := range resp.Device.Devices {
					if device.Status.Info != nil && device.Status.Info.PublicIP != nil {
						fmt.Println(*device.Status.Info.PublicIP)
					}
				}
				return nil
			}

			if showPublicKey {
				for _, device := range resp.Device.Devices {
					if device.Status.Info != nil && device.Status.Info.PublicIP != nil {
						fmt.Println(*device.Spec.PublicKey)
					}
				}
				return nil
			}

			if err := s.Printer.Print(resp.Device, printerConfig); err != nil {
				return err
			}
		}
	}
	if printerConfig.OutputFormat == print.JSON || printerConfig.OutputFormat == print.YAML {
		// raw
		var devices []api.Device
		var device api.Device
		if len(responseList) > 1 {
			for _, resp := range responseList {
				devices = append(devices, resp.Device.Devices...)
			}
			if err := s.Printer.Print(devices, printerConfig); err != nil {
				return err
			}
		} else {
			for _, resp := range responseList {
				if len(resp.Device.Devices) == 1 && len(ids) == 1 {
					device = resp.Device.Devices[0]
					if err := s.Printer.Print(device, printerConfig); err != nil {
						return err
					}
				} else {
					devices = resp.Device.Devices
					if err := s.Printer.Print(devices, printerConfig); err != nil {
						return err
					}
				}
			}

		}

	}

	return nil
}

func getDeviceBySN(cl *client.APIClient, sn string) (map[string]client.GetDevicesResponse, error) {
	return client.GetDevices(cl, sn, nil)
}

func getDevicesByDeploymentId(client *client.APIClient, deploymentID string) (api.Devices, error) {
	endpoint := fmt.Sprintf("%s?filter=deployment:%s", api.DevicesEndpoint, deploymentID)
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
