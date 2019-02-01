package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	api "github.com/grid-x/gxctl/pkg/api"
	client "github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	printer "github.com/grid-x/gxctl/pkg/printer"
)

type GetDeployments struct {
	Command *cobra.Command
}

func NewGetDeployments(parent *cobra.Command) *GetDeployments {
	var getDeploymentsCmd = &cobra.Command{
		Use:              "deployments",
		TraverseChildren: true,
		Aliases:          []string{"deployment", "deploy"},
		Short:            "get deployments",
		Long:             `Prints a list of all deployments you have access to`,
		RunE: func(cmd *cobra.Command, args []string) error {
			getCmdOutputType, _ := cmd.Flags().GetString("output")

			client := client.NewAPIClient()
			printer := printer.NewPrinter()

			if len(args) > 0 {
				//Get multiple deployments
				for _, a := range args {
					deployment, err := getDeploymentById(client, a)
					if err != nil {
						return err
					}

					err = printer.Print(deployment, getCmdOutputType)
					if err != nil {
						return err
					}
				}
			} else {
				//List all deployments
				deployments, err := getDeployments(client)
				if err != nil {
					return err
				}

				err = printer.Print(deployments, getCmdOutputType)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}

	parent.AddCommand(getDeploymentsCmd)

	return &GetDeployments{
		Command: getDeploymentsCmd,
	}
}

func getDeployments(client *client.APIClient) (api.Deployments, error) {
	response, err := client.GetRequest(api.DeploymentsEndpoint)
	if err != nil {
		return api.Deployments{}, err
	}

	deploymentList, err := api.NewDeployments(response)
	if err != nil {
		return deploymentList, err
	}

	if deploymentList.IsEmpty() {
		return deploymentList, errors.ListNotFoundError("deployments")
	}

	return deploymentList, nil
}

func getDeploymentById(client *client.APIClient, id string) (api.Deployment, error) {
	endpoint := fmt.Sprintf("%s/%s", api.DeploymentsEndpoint, id)
	response, err := client.GetRequest(endpoint)
	if err != nil {
		return api.Deployment{}, err
	}

	deployment, err := api.NewDeployment(response)
	if err != nil {
		return deployment, err
	}

	if deployment.IsEmpty() {
		msg := fmt.Sprintf("deployment \"%s\"", id)
		return deployment, errors.GetNotFoundError(msg)
	}

	return deployment, nil
}
