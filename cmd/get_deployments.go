package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	print "github.com/grid-x/gxctl/pkg/printer"
	"github.com/grid-x/gxctl/pkg/template"
)

type GetDeployments struct {
	Command *cobra.Command
}

func NewGetDeployments(parent *cobra.Command, client *client.APIClient, printer *print.Printer) *GetDeployments {
	var getDeploymentsCmd = &cobra.Command{
		Use:                   "deployment [ID] [OPTIONS]",
		DisableFlagsInUseLine: true,
		Aliases:               []string{"deployments", "deploy"},
		Short:                 "get deployment",
		Long:                  `Prints a list of all deployments you have access to`,
		Example:               "# Get all deployments \n  gxctl get deployments\n\n  # Get information about an deployment with abbreviation 3cc \n  gxctl get deployment 3cc",
		RunE: func(cmd *cobra.Command, args []string) error {
			getCmdOutputType, _ := cmd.Flags().GetString("output")

			printerConfig := print.Printconfig{
				OutputFormat: getCmdOutputType,
			}

			if len(args) > 0 {
				//Get multiple deployments
				//Lookup all existing deployments to validate ids and autocomplete them if necessary
				deployments, err := getDeployments(client)
				if err != nil {
					return err
				}
				deploymentIDs := deployments.GetIds()

				for _, a := range args {
					deployment, err := getDeploymentById(client, a, deploymentIDs)
					if err != nil {
						return err
					}

					if err := printer.Print(deployment, printerConfig); err != nil {
						return err
					}
				}
			} else {
				//List all deployments
				deployments, err := getDeployments(client)
				if err != nil {
					return err
				}

				if err := printer.Print(deployments, printerConfig); err != nil {
					return err
				}
			}
			return nil
		},
	}

	getDeploymentsCmd.SetHelpTemplate(template.HelpTemplate())
	getDeploymentsCmd.SetUsageTemplate(template.UsageTemplate())
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

func getDeploymentById(client *client.APIClient, id string, deploymentsIds []string) (api.Deployment, error) {
	deploymentID, err := api.LookupID(id, deploymentsIds)
	if err != nil {
		return api.Deployment{}, err
	}

	endpoint := fmt.Sprintf("%s/%s", api.DeploymentsEndpoint, deploymentID)
	response, err := client.GetRequest(endpoint)
	if err != nil {
		return api.Deployment{}, err
	}

	deployment, err := api.NewDeployment(response)
	if err != nil {
		return deployment, err
	}

	return deployment, nil
}
