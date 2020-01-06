package cmd

import (
	"fmt"

	deploymentsApi "github.com/grid-x/ds-api-types/management/2019-12-10/deployments"
	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	"github.com/grid-x/gxctl/pkg/template"
)

type UpdateDeployment struct {
	Command *cobra.Command
}

func NewUpdateDeployment(parent *cobra.Command, client *client.APIClient) *UpdateDevice {
	var updateDeploymentCmd = &cobra.Command{
		Use:                   "deployment ID [OPTIONS]",
		DisableFlagsInUseLine: true,
		Aliases:               []string{"deployments", "deploy"},
		Short:                 "update deployment",
		Long:                  `Updates a deployment`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.MissingParameter("ID", "gxctl update deployment -h")
			}

			if !api.IsDockerImageValid(args[0]) {
				return errors.InvalidParameter("IMAGE", "gxctl create deployment -h")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			updateDeploymentCmdImage, _ := cmd.Flags().GetString("image")
			updateDeploymentCmdApp, _ := cmd.Flags().GetString("app")
			updateDeploymentCmdSelector, _ := cmd.Flags().GetString("selector")

			if updateDeploymentCmdImage == "" && updateDeploymentCmdApp == "" && updateDeploymentCmdSelector == "" {
				return errors.NothingToDo("gxctl update deployment -h")
			}

			//Lookup all existing devices to validate ids and autocomplete them if necessary
			deployments, err := getDeployments(client)
			if err != nil {
				return err
			}
			deploymentIDs := deployments.GetIds()

			deployment, err := getDeploymentById(client, args[0], deploymentIDs)
			if err != nil {
				return err
			}

			if updateDeploymentCmdSelector != "" {
				matchByLabels, err := api.ParseMetadataMap(updateDeploymentCmdSelector)
				if err != nil {
					return err
				}

				deployment.Spec.Selector = deploymentsApi.Selector{
					MatchByLabels: matchByLabels,
				}
			}

			if updateDeploymentCmdImage != "" {
				name, err := api.GetDockerImageName(updateDeploymentCmdImage)
				if err != nil {
					return err
				}

				deployment.Spec.Template.Spec.Containers[0].Name = name
				deployment.Spec.Template.Spec.Containers[0].Image = updateDeploymentCmdImage
			}

			if updateDeploymentCmdApp != "" {
				deployment.Spec.App = updateDeploymentCmdApp
			}

			message, err := updateResource(client, deployment, deployment.Metadata.ID, nil)
			if err != nil {
				return err
			}

			fmt.Println(message)
			return nil
		},
	}

	updateDeploymentCmd.Flags().StringP("image", "i", "", "Image for the deployment")
	updateDeploymentCmd.Flags().StringP("app", "a", "", "App for the deployment")
	updateDeploymentCmd.Flags().StringP("selector", "s", "", "A space seperated list of labels eg. gridx.de/channel=stable gridx.de/area=west-1")

	updateDeploymentCmd.SetHelpTemplate(template.HelpTemplate())
	updateDeploymentCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(updateDeploymentCmd)

	return &UpdateDevice{
		Command: updateDeploymentCmd,
	}
}
