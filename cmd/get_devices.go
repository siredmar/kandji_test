package cmd

import (
	"fmt"
	"github.com/spf13/cobra"

	api "github.com/grid-x/gxctl/pkg/api"
	client "github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	printer "github.com/grid-x/gxctl/pkg/printer"
)

type GetDevices struct {
	Command *cobra.Command
}

func NewGetDevices(parent *cobra.Command) *GetDevices {
	var getDevicesCmd = &cobra.Command{
		Use:              "devices",
		TraverseChildren: true,
		Aliases:          []string{"device"},
		Short:            "get devices",
		Long:             `Prints a list of all devices you have access to`,
		RunE: func(cmd *cobra.Command, args []string) error {
			getCmdDeviceID, _ := cmd.Flags().GetString("device-id")
			getCmdOutputType, _ := cmd.Flags().GetString("output")

			client := client.NewAPIClient()
			printer := printer.NewPrinter()

			if len(args) > 0 {
				//Get multiple devices
				for _, a := range args {
					device, err := getDeviceById(client, a)
					if err != nil {
						return err
					}

					err = printer.Print(device, getCmdOutputType)
					if err != nil {
						return err
					}
				}
			} else if getCmdDeviceID != "" {
				//Get device
				device, err := getDeviceById(client, getCmdDeviceID)
				if err != nil {
					return err
				}

				err = printer.Print(device, getCmdOutputType)
				if err != nil {
					return err
				}
			} else {
				//List all devices
				devices, err := getDevices(client)
				if err != nil {
					return err
				}

				err = printer.Print(devices, getCmdOutputType)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}

	parent.AddCommand(getDevicesCmd)

	return &GetDevices{
		Command: getDevicesCmd,
	}
}

func getDevices(client *client.APIClient) (api.Devices, error) {
	response, err := client.Request(api.DevicesEndpoint)
	if err != nil {
		return api.Devices{}, err
	}

	deviceList, err := api.NewDevices(response)
	if err != nil {
		return deviceList, err
	}

	if len(deviceList.Devices) == 0 {
		return deviceList, errors.NotFoundError(errors.ErrorDetails{Command: "devices"})
	}

	return deviceList, nil
}

func getDeviceById(client *client.APIClient, id string) (api.Device, error) {
	endpoint := fmt.Sprintf("%s/%s", api.DevicesEndpoint, id)
	response, err := client.Request(endpoint)
	if err != nil {
		return api.Device{}, err
	}

	device, err := api.NewDevice(response)
	if err != nil {
		return device, err
	}

	if device.ID != id {
		return device, errors.NotFoundError(errors.ErrorDetails{Command: "device", Id: id})
	}

	return device, nil
}
