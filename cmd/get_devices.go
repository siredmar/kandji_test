package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	print "github.com/grid-x/gxctl/pkg/printer"
	"github.com/grid-x/gxctl/pkg/template"
)

type GetDevices struct {
	Command *cobra.Command
}

func NewGetDevices(parent *cobra.Command, client *client.APIClient, printer *print.Printer) *GetDevices {
	var getDevicesCmd = &cobra.Command{
		Use:                   "device [ID] [OPTIONS]",
		DisableFlagsInUseLine: true,
		Aliases:               []string{"devices"},
		Short:                 "get device",
		Example:               "# Get all devices \n  gxctl get devices\n\n  # Get information about an device with abbreviation c72 \n  gxctl get device c72",
		Long:                  `Prints a list of all devices you have access to`,
		Args: func(cmd *cobra.Command, args []string) error {
			getCmdShowDockerconfig, _ := cmd.Flags().GetBool("show-dockerconfig")
			getCmdShowPods, _ := cmd.Flags().GetBool("show-pods")
			getCmdShowPublickey, _ := cmd.Flags().GetBool("show-publickey")

			if getCmdShowDockerconfig && len(args) != 1 {
				return errors.InvalidParameter("show-dockerconfig", "docker-configs can just be shown for a single device")
			}

			if getCmdShowPods && len(args) != 1 {
				return errors.InvalidParameter("show-pods", "pods can just be shown for a single device")
			}

			if getCmdShowPublickey && len(args) != 1 {
				return errors.InvalidParameter("show-publickey", "publickeys can just be shown for a single device")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			getCmdOutputType, _ := cmd.Flags().GetString("output")
			getCmdSortBy, _ := cmd.Flags().GetString("sort-by")
			getCmdShowDockerconfig, _ := cmd.Flags().GetBool("show-dockerconfig")
			getCmdShowPods, _ := cmd.Flags().GetBool("show-pods")
			getCmdShowPublickey, _ := cmd.Flags().GetBool("show-publickey")
			getCmdShowAll, _ := cmd.Flags().GetBool("all")

			printerConfig := print.Printconfig{
				OutputFormat: getCmdOutputType,
				SortBy:       getCmdSortBy,
				ShowAll:      getCmdShowAll,
			}

			if len(args) > 0 {
				//Get multiple devices
				//Lookup all existing devices to validate ids and autocomplete them if necessary
				devices, err := getDevices(client)
				if err != nil {
					return err
				}
				deviceIDs := devices.GetIds()

				for _, a := range args {
					if getCmdShowPublickey {
						// Just show publickey
						device, err := getDeviceById(client, a, deviceIDs)
						if err != nil {
							return err
						}
						fmt.Println(*device.Spec.PublicKey)
					} else if getCmdShowDockerconfig {
						// Just show dockerconfig
						configs, err := getDeviceDockerConfigs(client, a, deviceIDs)
						if err != nil {
							return err
						}

						if err := printer.Print(configs, printerConfig); err != nil {
							return err
						}
					} else if getCmdShowPods {
						// Just show pods
						pods, err := getDevicePods(client, a, deviceIDs)
						if err != nil {
							return err
						}

						if err := printer.Print(pods, printerConfig); err != nil {
							return err
						}
					} else {
						// Print device
						device, err := getDeviceById(client, a, deviceIDs)
						if err != nil {
							return err
						}

						if err := printer.Print(device, printerConfig); err != nil {
							return err
						}
					}
				}
			} else {
				//List all devices
				devices, err := getDevices(client)
				if err != nil {
					return err
				}

				if err := printer.Print(devices, printerConfig); err != nil {
					return err
				}
			}
			return nil
		},
	}
	getDevicesCmd.Flags().BoolP("show-dockerconfig", "", false, "print the docker config for a device")
	getDevicesCmd.Flags().BoolP("show-pods", "", false, "print the docker config for a device")
	getDevicesCmd.Flags().BoolP("show-publickey", "", false, "print the public key for a device")
	getDevicesCmd.Flags().BoolP("all", "", false, "show also inactive devices")
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

	configs, err := api.NewDeviceDockerConfigs(response)
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

	pods, err := api.NewPods(response)
	if err != nil {
		return pods, err
	}

	return pods, nil
}
