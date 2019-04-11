package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	"github.com/grid-x/gxctl/pkg/template"
)

type PatchDevice struct {
	Command *cobra.Command
}

func NewPatchDevice(parent *cobra.Command, client *client.APIClient) *PatchDevice {
	var patchDeviceCmd = &cobra.Command{
		Use:                   "device ID [OPTIONS]",
		DisableFlagsInUseLine: true,
		Short:                 "patch device",
		Long:                  `Patches a device`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.MissingParameter("ID", "gxctl patch device -h")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			patchDeviceCmdMaintenanceWindow, _ := cmd.Flags().GetString("maintenance-window")
			patchDeviceCmdMacAddress, _ := cmd.Flags().GetString("mac-address")
			patchDeviceCmdLabels, _ := cmd.Flags().GetString("labels")

			if patchDeviceCmdMaintenanceWindow == "" && patchDeviceCmdMacAddress == "" && patchDeviceCmdLabels == "" {
				return errors.NothingToDo("gxctl patch device -h")
			}

			//Lookup all existing devices to validate ids and autocomplete them if necessary
			devices, err := getDevices(client)
			if err != nil {
				return err
			}
			deviceIDs := devices.GetIds()

			d := api.PatchDevice{}

			if patchDeviceCmdMaintenanceWindow != "" {
				d.Spec.MaintenanceWindow = &patchDeviceCmdMaintenanceWindow
			}
			if patchDeviceCmdMacAddress != "" {
				d.Spec.MACAddress = &patchDeviceCmdMacAddress
			}

			if patchDeviceCmdLabels != "" {
				labels := strings.Split(patchDeviceCmdLabels, " ")
				m := make(map[string]string)
				for _, pair := range labels {
					z := strings.Split(pair, "=")
					m[z[0]] = z[1]
				}

				d.Metadata.Labels = m
			}
			message, err := patchResource(client, d, args[0], deviceIDs)
			if err != nil {
				return err
			}

			fmt.Println(message)
			return nil
		},
	}

	patchDeviceCmd.Flags().StringP("maintenance-window", "m", "", "Maintenance window for the device")
	patchDeviceCmd.Flags().StringP("mac-address", "a", "", "Mac address for the device")
	patchDeviceCmd.Flags().StringP("labels", "l", "", "A space seperated list of labels eg. gridx.de/channel=stable gridx.de/area=west-1")

	patchDeviceCmd.SetHelpTemplate(template.HelpTemplate())
	patchDeviceCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(patchDeviceCmd)

	return &PatchDevice{
		Command: patchDeviceCmd,
	}
}
