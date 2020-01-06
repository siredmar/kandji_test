package cmd

import (
	"fmt"

	types "github.com/grid-x/ds-api-types"
	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	"github.com/grid-x/gxctl/pkg/template"
)

type UpdateDevice struct {
	Command *cobra.Command
}

func NewUpdateDevice(parent *cobra.Command, client *client.APIClient) *UpdateDevice {
	var updateDeviceCmd = &cobra.Command{
		Use:                   "device ID [OPTIONS]",
		DisableFlagsInUseLine: true,
		Short:                 "update device",
		Long:                  `Updates a device`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.MissingParameter("ID", "gxctl update device -h")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			updateDeviceCmdMaintenanceWindow, _ := cmd.Flags().GetString("maintenance-window")
			updateDeviceCmdMacAddress, _ := cmd.Flags().GetString("mac-address")
			updateDeviceCmdLabels, _ := cmd.Flags().GetString("labels")
			updateDeviceCmdAnnotations, _ := cmd.Flags().GetString("annotations")

			if updateDeviceCmdMaintenanceWindow == "" && updateDeviceCmdMacAddress == "" && updateDeviceCmdLabels == "" && updateDeviceCmdAnnotations == "" {
				return errors.NothingToDo("gxctl update device -h")
			}

			//Lookup all existing devices to validate ids and autocomplete them if necessary
			devices, err := getDevices(client)
			if err != nil {
				return err
			}
			deviceIDs := devices.GetIds()

			d, err := getDeviceById(client, args[0], deviceIDs)
			if err != nil {
				return err
			}

			if updateDeviceCmdMaintenanceWindow != "" {
				w, err := types.NewMaintenanceWindow(updateDeviceCmdMaintenanceWindow)
				if err != nil {
					return err
				}
				d.Spec.MaintenanceWindow = w
			}
			if updateDeviceCmdMacAddress != "" {
				d.Spec.MACAddress = &updateDeviceCmdMacAddress
			}

			if updateDeviceCmdLabels != "" {
				m, err := api.ParseMetadataMap(updateDeviceCmdLabels)
				if err != nil {
					return err
				}

				d.Metadata.Labels = m
			}

			if updateDeviceCmdAnnotations != "" {
				a, err := api.ParseMetadataMap(updateDeviceCmdAnnotations)
				if err != nil {
					return err
				}

				d.Metadata.Annotations = a
			}

			message, err := updateResource(client, d, d.Metadata.ID, nil)
			if err != nil {
				return err
			}

			fmt.Println(message)
			return nil
		},
	}

	updateDeviceCmd.Flags().StringP("maintenance-window", "w", "", "Maintenance window for the device")
	updateDeviceCmd.Flags().StringP("mac-address", "m", "", "Mac address for the device")
	updateDeviceCmd.Flags().StringP("labels", "l", "", "A space seperated list of labels eg. gridx.de/channel=stable gridx.de/area=west-1")
	updateDeviceCmd.Flags().StringP("annotations", "a", "", "A space seperated list of annotations eg. gridx.ai/custimer=123 gridx.ai/style=red")

	updateDeviceCmd.SetHelpTemplate(template.HelpTemplate())
	updateDeviceCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(updateDeviceCmd)

	return &UpdateDevice{
		Command: updateDeviceCmd,
	}
}
