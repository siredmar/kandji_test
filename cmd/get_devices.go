package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	api "github.com/grid-x/gxctl/pkg/api"
	client "github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	printer "github.com/grid-x/gxctl/pkg/printer"
	template "github.com/grid-x/gxctl/pkg/template"
)

type GetDevices struct {
	Command *cobra.Command
}

func NewGetDevices(parent *cobra.Command) *GetDevices {
	var getDevicesCmd = &cobra.Command{
		Use:                   "device [ID] [OPTIONS]",
		DisableFlagsInUseLine: true,
		Aliases:               []string{"devices"},
		Short:                 "get device",
		Example:               "# Get all devices \n  gxctl get devices\n\n  # Get information about an device with abbreviation c72 \n  gxctl get device c72",
		Long:                  `Prints a list of all devices you have access to`,
		RunE: func(cmd *cobra.Command, args []string) error {
			getCmdOutputType, _ := cmd.Flags().GetString("output")

			client := client.NewAPIClient()
			printer := printer.NewPrinter()

			if len(args) > 0 {
				//Get multiple devices
				//Lookup all existing devices to validate ids and autocomplete them if necessary
				devices, err := getDevices(client)
				if err != nil {
					return err
				}
				deviceIDs := devices.GetIds()

				for _, a := range args {
					device, err := getDeviceById(client, a, deviceIDs)
					if err != nil {
						return err
					}

					err = printer.Print(device, getCmdOutputType)
					if err != nil {
						return err
					}
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

	getDevicesCmd.SetHelpTemplate(template.HelpTemplate())
	getDevicesCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(getDevicesCmd)

	return &GetDevices{
		Command: getDevicesCmd,
	}
}

func getDevices(client *client.APIClient) (api.Devices, error) {
	response, err := client.GetRequest(api.DevicesEndpoint)
	if err != nil {
		return api.Devices{}, err
	}

	deviceList, err := api.NewDevices(response)
	if err != nil {
		return deviceList, err
	}

	if deviceList.IsEmpty() {
		return deviceList, errors.ListNotFoundError("devices")
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

	device, err := api.NewDevice(response)
	if err != nil {
		return device, err
	}

	return device, nil
}
