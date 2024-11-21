package action

import (
	"fmt"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/service"
)

func DeleteMaintenance(s *service.Service, ids []string) error {
	// Lookup all existing deployments to validate ids and autocomplete them if necessary
	tasks, err := getMaintenanceTasks(s.Client)
	if err != nil {
		return err
	}
	taskIDs := tasks.GetIDs()

	for _, id := range ids {
		msg, err := deleteMaintenanceTask(s.Client, id, taskIDs)
		if err != nil {
			return err
		}
		fmt.Println(msg)
	}
	return nil
}

func deleteMaintenanceTask(client *client.APIClient, id string, taskIDs []string) (string, error) {
	taskID, err := api.LookupID(id, taskIDs)
	if err != nil {
		return "", err
	}

	_, err = client.DeleteRequest(api.MaintenanceEndpoint, taskID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Maintenance task %s deleted successfully", taskID), nil
}
