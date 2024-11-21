package action

import (
	"fmt"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
	"github.com/grid-x/gxctl/pkg/printer"
	"github.com/grid-x/gxctl/pkg/service"
)

func GetMaintenance(s *service.Service, outputType string, ids []string) error {
	printerConfig := printer.PrintConfig{
		OutputFormat: outputType,
	}

	if len(ids) > 0 {
		// Get multiple maintenance tasks
		// Lookup all existing maintenance task to validate ids and autocomplete them if necessary
		tasks, err := getMaintenanceTasks(s.Client)
		if err != nil {
			return err
		}
		taskIDs := tasks.GetIDs()

		for _, a := range ids {
			task, err := getMaintenanceTaskByID(s.Client, a, taskIDs)
			if err != nil {
				return err
			}

			if err := s.Printer.Print(task, printerConfig); err != nil {
				return err
			}
		}
	} else {
		// List all deployments
		tasks, err := getMaintenanceTasks(s.Client)
		if err != nil {
			return err
		}

		if err := s.Printer.Print(tasks, printerConfig); err != nil {
			return err
		}
	}

	return nil
}

func getMaintenanceTasks(client *client.APIClient) (api.MaintenanceTasks, error) {
	response, err := client.GetRequest(api.MaintenanceEndpoint)
	if err != nil {
		return api.MaintenanceTasks{}, err
	}

	taskList, err := api.NewMaintenanceTasks(response, false)
	if err != nil {
		return taskList, err
	}

	if taskList.IsEmpty() {
		return taskList, errors.E(
			errors.NotExists,
			"no maintenance tasks found",
		)
	}

	return taskList, nil
}

func getMaintenanceTaskByID(client *client.APIClient, id string, taskIDs []string) (api.MaintenanceTask, error) {
	taskID, err := api.LookupID(id, taskIDs)
	if err != nil {
		return api.MaintenanceTask{}, err
	}

	endpoint := fmt.Sprintf("%s/%s", api.MaintenanceEndpoint, taskID)
	response, err := client.GetRequest(endpoint)
	if err != nil {
		return api.MaintenanceTask{}, err
	}

	task, err := api.NewMaintenanceTask(response, false)
	if err != nil {
		return task, err
	}

	return task, nil
}
