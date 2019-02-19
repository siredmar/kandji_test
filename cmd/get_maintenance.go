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

type GetMaintenance struct {
	Command *cobra.Command
}

func NewGetMaintenance(parent *cobra.Command) *GetMaintenance {
	var getMaintenanceCmd = &cobra.Command{
		Use:     "maintenance",
		Short:   "Get different maintenances",
		Aliases: []string{"maintenances"},
		Long:    `TODO`,
		Example: "# Get all maintenance tasks \n  gxctl get maintenances\n\n  # Get information about an maintenance task with abbreviation a56 \n  gxctl get maintenance a56",
		RunE: func(cmd *cobra.Command, args []string) error {
			getCmdOutputType, _ := cmd.Flags().GetString("output")

			client := client.NewAPIClient()
			printer := printer.NewPrinter()

			if len(args) > 0 {
				//Get multiple maintenance tasks
				//Lookup all existing maintenance task to validate ids and autocomplete them if necessary
				tasks, err := getMaintenanceTasks(client)
				if err != nil {
					return err
				}
				taskIDs := tasks.GetIds()

				for _, a := range args {
					task, err := getMaintenanceTaskById(client, a, taskIDs)
					if err != nil {
						return err
					}

					err = printer.Print(task, getCmdOutputType)
					if err != nil {
						return err
					}
				}
			} else {
				//List all deployments
				tasks, err := getMaintenanceTasks(client)
				if err != nil {
					return err
				}

				err = printer.Print(tasks, getCmdOutputType)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}

	getMaintenanceCmd.SetHelpTemplate(template.HelpTemplate())
	getMaintenanceCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(getMaintenanceCmd)

	return &GetMaintenance{
		Command: getMaintenanceCmd,
	}
}

func getMaintenanceTasks(client *client.APIClient) (api.MaintenanceTasks, error) {
	response, err := client.GetRequest(api.MaintenanceEndpoint)
	if err != nil {
		return api.MaintenanceTasks{}, err
	}

	taskList, err := api.NewMaintenanceTasks(response)
	if err != nil {
		return taskList, err
	}

	if taskList.IsEmpty() {
		return taskList, errors.ListNotFoundError("maintenance tasks")
	}

	return taskList, nil
}

func getMaintenanceTaskById(client *client.APIClient, id string, taskIds []string) (api.MaintenanceTask, error) {
	taskId, err := api.LookupID(id, taskIds)
	if err != nil {
		return api.MaintenanceTask{}, err
	}

	endpoint := fmt.Sprintf("%s/%s", api.MaintenanceEndpoint, taskId)
	response, err := client.GetRequest(endpoint)
	if err != nil {
		return api.MaintenanceTask{}, err
	}

	task, err := api.NewMaintenanceTask(response)
	if err != nil {
		return task, err
	}

	return task, nil
}
