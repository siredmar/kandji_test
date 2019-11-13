package cmd

import (
	"fmt"

	maintenanceApi "github.com/grid-x/ds-api-types/management/2019-11-04/maintenance"
	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	"github.com/grid-x/gxctl/pkg/template"
)

type Restart struct {
	Command *cobra.Command
}

func NewRestart(parent *cobra.Command, client *client.APIClient) *Restart {
	var restartCmd = &cobra.Command{
		Use:                   "restart ID [OPTIONS]",
		DisableFlagsInUseLine: true,
		Short:                 "restart different devices",
		Long:                  `TODO`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.MissingParameter("ID", "gxctl restart -h")
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

			message, err := restartDevice(client, args[0], deviceIDs)
			if err != nil {
				return err
			}

			fmt.Println(message)
			return nil
		},
	}

	restartCmd.SetHelpTemplate(template.HelpTemplate())
	restartCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(restartCmd)

	return &Restart{
		Command: restartCmd,
	}
}

func restartDevice(client *client.APIClient, id string, ids []string) (string, error) {
	devID := id
	if ids != nil {
		var err error
		devID, err = api.LookupID(id, ids)
		if err != nil {
			return "", err
		}
	}

	d := api.CreateMaintenanceTask{
		Spec: &maintenanceApi.MaintenanceTaskSpec{
			Type:     maintenanceApi.MaintenanceTaskTypeRestart,
			DeviceID: devID,
		},
	}

	message, err := createResource(client, d)
	if err != nil {
		return "", err
	}

	return message, nil
}
