package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
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
		Args: func(cmd *cobra.Command, args []string) error {
			getCmdShowDevices, _ := cmd.Flags().GetBool("show-devices")

			if getCmdShowDevices && len(args) != 1 {
				return errors.E(
					errors.Invalid,
					"devices can just be shown for a single deployment",
					[]string{"run 'gxctl get deployment --help' for usage"},
				)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			getCmdOutputType, _ := cmd.Flags().GetString("output")
			getCmdSortBy, _ := cmd.Flags().GetString("sort-by")
			getCmdShowDevices, _ := cmd.Flags().GetBool("show-devices")

			printerConfig := print.Printconfig{
				OutputFormat: getCmdOutputType,
				SortBy:       getCmdSortBy,
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
					if getCmdShowDevices {
						// Just show devices
						devices, err := getDeploymentDevices(client, a, deploymentIDs)
						if err != nil {
							return err
						}

						if err := printer.Print(devices, printerConfig); err != nil {
							return err
						}
					} else {
						deployment, err := getDeploymentById(client, a, deploymentIDs)
						if err != nil {
							return err
						}

						if err := printer.Print(deployment, printerConfig); err != nil {
							return err
						}
					}
				}
			} else {
				deployments, err := getDeployments(client)
				if err != nil {
					return err
				}

				var label []api.Deployment
				var id []api.Deployment

				for _, d := range deployments.Deployments {
					if d.Spec.Selector.MatchByDeviceID != nil {
						id = append(id, d)
						continue
					}
					label = append(label, d)
				}

				fmt.Print("Deployments by ID\n\n")
				if err := printer.Print(api.Deployments{Deployments: id}, printerConfig); err != nil {
					return err
				}

				fmt.Print("Deployments by Label\n\n")
				if err := printer.Print(api.Deployments{Deployments: label}, printerConfig); err != nil {
					return err
				}
			}
			return nil
		},
	}

	getDeploymentsCmd.Flags().BoolP("show-devices", "", false, "print the devices for a deployment")
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
		return deploymentList, errors.E(
			errors.NotExists,
			"no deployments found",
		)
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

func getDeploymentDevices(client *client.APIClient, id string, deploymentIds []string) (api.Devices, error) {
	deploymentID, err := api.LookupID(id, deploymentIds)
	if err != nil {
		return api.Devices{}, err
	}

	endpoint := fmt.Sprintf("%s/%s/devices", api.DeploymentsEndpoint, deploymentID)
	response, err := client.GetRequest(endpoint)
	if err != nil {
		return api.Devices{}, err
	}

	devices, err := api.NewDevices(response)
	if err != nil {
		return devices, err
	}

	return devices, nil
}
