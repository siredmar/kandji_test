package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	api "github.com/grid-x/gxctl/pkg/api"
	client "github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	template "github.com/grid-x/gxctl/pkg/template"
)

type DeleteMaintenance struct {
	Command *cobra.Command
}

func NewDeleteMaintenance(parent *cobra.Command) *DeleteMaintenance {
	var deleteMaintenanceCmd = &cobra.Command{
		Use:              "maintenance ID [OPTIONS]",
		TraverseChildren: true,
		Aliases:          []string{"maintenances"},
		Short:            "delete maintenance",
		Long:             `TODO`,
		Example:          "# Delete a maintenance task \n  gxctl delete maintenance 3409845a-75d1-452a-ab81-501c225c1d1b",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.MissingParameter("ID", "gxctl delete maintenance -h")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client := client.NewAPIClient()

			//Lookup all existing deployments to validate ids and autocomplete them if necessary
			tasks, err := getMaintenanceTasks(client)
			if err != nil {
				return err
			}
			taskIds := tasks.GetIds()

			for _, id := range args {
				msg, err := deleteMaintenanceTask(client, id, taskIds)
				if err != nil {
					return err
				}
				fmt.Println(msg)
			}
			return nil
		},
	}

	deleteMaintenanceCmd.SetHelpTemplate(template.HelpTemplate())
	deleteMaintenanceCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(deleteMaintenanceCmd)

	return &DeleteMaintenance{
		Command: deleteMaintenanceCmd,
	}
}

func deleteMaintenanceTask(client *client.APIClient, id string, taskIds []string) (string, error) {
	taskId, err := api.LookupID(id, taskIds)
	if err != nil {
		return "", err
	}

	_, err = client.DeleteRequest(api.MaintenanceEndpoint, taskId)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Maintenance task %s deleted successfully", taskId), nil
}
