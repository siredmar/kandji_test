package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	"github.com/grid-x/gxctl/pkg/template"
)

type DeleteDevice struct {
	Command *cobra.Command
}

func NewDeleteDevice(parent *cobra.Command, client *client.APIClient) *DeleteApplication {
	var deleteDeviceCmd = &cobra.Command{
		Use:                   "device ID",
		Aliases:               []string{"devices"},
		DisableFlagsInUseLine: true,
		Short:                 "delete device",
		Long:                  `TODO`,
		Example:               "# Delete devices with abbreviation a42\n  gxctl delete device a42",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.MissingParameter("NAME", "gxctl delete device -h")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, d := range args {
				msg, err := deleteDevice(client, d)
				if err != nil {
					return err
				}
				fmt.Println(msg)
			}
			return nil
		},
	}

	deleteDeviceCmd.SetHelpTemplate(template.HelpTemplate())
	deleteDeviceCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(deleteDeviceCmd)

	return &DeleteApplication{
		Command: deleteDeviceCmd,
	}
}

func deleteDevice(client *client.APIClient, deviceID string) (string, error) {
	_, err := client.DeleteRequest(api.DevicesEndpoint, deviceID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Device %s deleted successfully", deviceID), nil
}
