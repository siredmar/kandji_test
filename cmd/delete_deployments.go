package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	"github.com/grid-x/gxctl/pkg/template"
)

type DeleteDeployment struct {
	Command *cobra.Command
}

func NewDeleteDeployment(parent *cobra.Command, client *client.APIClient) *DeleteDeployment {
	var deleteDeploymentCmd = &cobra.Command{
		Use:              "deployment ID [OPTIONS]",
		TraverseChildren: true,
		Aliases:          []string{"deployments", "deploy"},
		Short:            "delete deployment",
		Long:             `TODO`,
		Example:          "# Delete a deployment \n  gxctl delete deployment 337df243-2cc9-46f4-bfeb-3c978ece4252",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.MissingParameter("ID", "gxctl delete deployment -h")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			//Lookup all existing deployments to validate ids and autocomplete them if necessary
			deployments, err := getDeployments(client)
			if err != nil {
				return err
			}
			deploymentIDs := deployments.GetIds()

			for _, id := range args {
				msg, err := deleteDeployment(client, id, deploymentIDs)
				if err != nil {
					return err
				}
				fmt.Println(msg)
			}
			return nil
		},
	}

	deleteDeploymentCmd.SetHelpTemplate(template.HelpTemplate())
	deleteDeploymentCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(deleteDeploymentCmd)

	return &DeleteDeployment{
		Command: deleteDeploymentCmd,
	}
}

func deleteDeployment(client *client.APIClient, id string, deploymentsIds []string) (string, error) {
	deploymentID, err := api.LookupID(id, deploymentsIds)
	if err != nil {
		return "", err
	}

	_, err = client.DeleteRequest(api.DeploymentsEndpoint, deploymentID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Deployment %s deleted successfully", deploymentID), nil
}
