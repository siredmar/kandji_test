package cmd

import (
	"fmt"
	"strings"

	deploymentsApi "github.com/grid-x/ds-api-types/management/2018-11-28/deployments"
	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	"github.com/grid-x/gxctl/pkg/template"
)

type PatchDeployment struct {
	Command *cobra.Command
}

func NewPatchDeployment(parent *cobra.Command, client *client.APIClient) *PatchDevice {
	var patchDeploymentCmd = &cobra.Command{
		Use:                   "deployment ID [OPTIONS]",
		DisableFlagsInUseLine: true,
		Aliases:               []string{"deployments", "deploy"},
		Short:                 "patch deployment",
		Long:                  `Patches a deployment`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.MissingParameter("ID", "gxctl patch deployment -h")
			}

			if !api.IsDockerImageValid(args[0]) {
				return errors.InvalidParameter("IMAGE", "gxctl create deployment -h")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			patchDeploymentCmdImage, _ := cmd.Flags().GetString("image")
			patchDeploymentCmdApp, _ := cmd.Flags().GetString("app")
			patchDeploymentCmdSelector, _ := cmd.Flags().GetString("selector")

			if patchDeploymentCmdImage == "" && patchDeploymentCmdApp == "" && patchDeploymentCmdSelector == "" {
				return errors.NothingToDo("gxctl patch deployment -h")
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

			if patchDeploymentCmdSelector != "" {
				matchByLabels := make(map[string]string)
				labels := strings.Split(patchDeploymentCmdSelector, " ")
				for _, pair := range labels {
					z := strings.Split(pair, "=")
					matchByLabels[z[0]] = z[1]
				}

				deployment.Spec.Selector = deploymentsApi.Selector{
					MatchByLabels: matchByLabels,
				}
			}

			if patchDeploymentCmdImage != "" {
				name, err := api.GetDockerImageName(patchDeploymentCmdImage)
				if err != nil {
					return err
				}

				deployment.Spec.Template.Spec.Containers[0].Name = name
				deployment.Spec.Template.Spec.Containers[0].Image = patchDeploymentCmdImage
			}
			if patchDeploymentCmdApp != "" {
				deployment.Spec.App = patchDeploymentCmdApp
			}

			d := api.PatchDeployment{}
			d.Spec = &deployment.Spec

			message, err := patchResource(client, d, deployment.Metadata.ID, nil)
			if err != nil {
				return err
			}

			fmt.Println(message)
			return nil
		},
	}

	patchDeploymentCmd.Flags().StringP("image", "i", "", "Image for the deployment")
	patchDeploymentCmd.Flags().StringP("app", "a", "", "App for the deployment")
	patchDeploymentCmd.Flags().StringP("selector", "s", "", "A space seperated list of labels eg. gridx.de/channel=stable gridx.de/area=west-1")

	patchDeploymentCmd.SetHelpTemplate(template.HelpTemplate())
	patchDeploymentCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(patchDeploymentCmd)

	return &PatchDevice{
		Command: patchDeploymentCmd,
	}
}
