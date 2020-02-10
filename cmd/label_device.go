package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
	"github.com/grid-x/gxctl/pkg/template"
)

type LabelDevice struct {
	Command *cobra.Command
}

func NewLabelDevice(parent *cobra.Command, client *client.APIClient) *LabelDevice {
	var labelDeviceCmd = &cobra.Command{
		Use:                   "device ID KEY_1=VAL_1 ... KEY_N=VAL_N [OPTIONS]",
		DisableFlagsInUseLine: true,
		Aliases:               []string{"devices"},
		Short:                 "label device",
		Long:                  `Todo`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.E(
					errors.Invalid,
					"required argument ID not found",
					[]string{"run 'gxctl label device --help' for usage"},
				)
			}
			if len(args) == 1 {
				return errors.E(
					errors.Invalid,
					"required arguments KEY=VAL not found",
					[]string{"run 'gxctl label device --help' for usage"},
				)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
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

			m, err := api.ParseMetadataMap(strings.Join(args[1:], " "))
			if err != nil {
				return err
			}

			d.Metadata.Labels = m

			message, err := updateResource(client, d, d.Metadata.ID, nil)
			if err != nil {
				return err
			}

			fmt.Println(message)
			return nil
		},
	}

	labelDeviceCmd.SetHelpTemplate(template.HelpTemplate())
	labelDeviceCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(labelDeviceCmd)

	return &LabelDevice{
		Command: labelDeviceCmd,
	}
}
