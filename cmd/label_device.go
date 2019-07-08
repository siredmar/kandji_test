package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	types "github.com/grid-x/ds-api-types"
	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
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
				return errors.MissingParameter("ID", "gxctl create application -h")
			}
			if len(args) == 1 {
				return errors.MissingParameter("LABEL", "gxctl create application -h")
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

			d := api.PatchDevice{}

			m, err := api.ParseMetadataMap(strings.Join(args[1:], " "))
			if err != nil {
				return err
			}

			if d.Metadata == nil {
				d.Metadata = &types.UpdateMetadata{}
			}

			d.Metadata.Labels = m

			message, err := patchResource(client, d, args[0], deviceIDs)
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
