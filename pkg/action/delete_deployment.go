package action

import (
	"fmt"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/service"
)

func DeleteDeployment(s *service.Service, ids []string) error {
	// Lookup all existing deployments to validate ids and autocomplete them if necessary
	deployments, err := getDeployments(s.Client)
	if err != nil {
		return err
	}
	deploymentIDs := deployments.GetIDs()

	for _, id := range ids {
		msg, err := deleteDeployment(s.Client, id, deploymentIDs)
		if err != nil {
			return err
		}
		fmt.Println(msg)
	}
	return nil
}

func deleteDeployment(client *client.APIClient, id string, deploymentsIDs []string) (string, error) {
	deploymentID, err := api.LookupID(id, deploymentsIDs)
	if err != nil {
		return "", err
	}

	_, err = client.DeleteRequest(api.DeploymentsEndpoint, deploymentID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Deployment %s deleted successfully", deploymentID), nil
}
