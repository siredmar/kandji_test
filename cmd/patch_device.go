package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/cobra"

	api "github.com/grid-x/gxctl/pkg/api"
	client "github.com/grid-x/gxctl/pkg/client"
)

type PatchDevice struct {
	Command *cobra.Command
}

func NewPatchDevice(parent *cobra.Command) *PatchDevice {
	var patchDeviceCmd = &cobra.Command{
		Use:              "device",
		TraverseChildren: true,
		Short:            "patch device",
		Long:             `Patches a device`,
		RunE: func(cmd *cobra.Command, args []string) error {
			patchDeviceCmdFilename, _ := cmd.Flags().GetString("filename")
			patchDeviceCmdMaintenanceWindow, _ := cmd.Flags().GetString("maintenance-window")
			patchDeviceCmdMacAddress, _ := cmd.Flags().GetString("mac-address")
			patchDeviceCmdLabels, _ := cmd.Flags().GetString("labels")

			if patchDeviceCmdFilename != "" && (patchDeviceCmdMaintenanceWindow != "" || patchDeviceCmdMacAddress != "" || patchDeviceCmdLabels != "") {
				return errors.New("You can either specify a patchfile or maintenance-window/mac-address/labels but not both")
			}
			if patchDeviceCmdFilename == "" && patchDeviceCmdMaintenanceWindow == "" && patchDeviceCmdMacAddress == "" && patchDeviceCmdLabels == "" {
				cmd.Usage()
				return errors.New("Nothing to do")
			}

			client := client.NewAPIClient()

			if len(args) != 1 {
				return errors.New("Missing device ID")
			}

			var res interface{}

			if patchDeviceCmdFilename != "" {
				jsonFile, err := os.Open(patchDeviceCmdFilename)
				if err != nil {
					return err
				}
				defer jsonFile.Close()

				bytes, err := ioutil.ReadAll(jsonFile)
				if err != nil {
					return err
				}

				res, err = checkPatchResourceFile(bytes)
				if err != nil {
					return err
				}

			} else {
				d := api.PatchDevice{}

				if patchDeviceCmdMaintenanceWindow != "" {
					d.Spec.MaintenanceWindow = &patchDeviceCmdMaintenanceWindow
				}
				if patchDeviceCmdMacAddress != "" {
					d.Spec.MACAddress = &patchDeviceCmdMacAddress
				}

				if patchDeviceCmdLabels != "" {
					labels := strings.Split(patchDeviceCmdLabels, ",")
					m := make(map[string]string)
					for _, pair := range labels {
						z := strings.Split(pair, ":")
						m[z[0]] = z[1]
					}

					d.Metadata.Labels = m
				}

				res = d
			}

			message, err := patchResource(client, res, args[0])
			if err != nil {
				return err
			}

			fmt.Println(message)
			return nil
		},
	}

	patchDeviceCmd.Flags().StringP("maintenance-window", "m", "", "Maintenance window for the device")
	patchDeviceCmd.Flags().StringP("mac-address", "a", "", "Mac address for the device")
	patchDeviceCmd.Flags().StringP("labels", "l", "", "A comma seperated list of labels eg. gridx.de/channel:stable,gridx.de/area:west-1")
	parent.AddCommand(patchDeviceCmd)

	return &PatchDevice{
		Command: patchDeviceCmd,
	}
}
