package cmd

import (
	"fmt"

	podsApi "github.com/grid-x/ds-api-types/management/2019-08-17/pod"
	deploymentsApi "github.com/grid-x/ds-api-types/management/2019-12-10/deployments"
	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
	"github.com/grid-x/gxctl/pkg/template"
)

type CreateDeployment struct {
	Command *cobra.Command
}

func NewCreateDeployment(parent *cobra.Command, client *client.APIClient) *CreateDeployment {
	var createDeploymentCmd = &cobra.Command{
		Use:                   "deployment IMAGE --app=APP --selector=SELECTOR [OPTIONS]",
		Short:                 "Creates an deployment",
		DisableFlagsInUseLine: true,
		Long:                  `TODO`,
		Aliases:               []string{"deployments", "deploy"},
		Example:               "# Create an nginx deployment for app test \n  gxctl create deployment nginx:1.15.8 -a test -s gridx.de/channel=stable",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.E(
					errors.Invalid,
					"required argument IMAGE not found",
					[]string{"run 'gxctl create deployment --help' for usage"},
				)
			}

			if !api.IsDockerImageValid(args[0]) {
				return errors.E(
					errors.Invalid,
					"required argument IMAGE is invalid",
					[]string{"run 'gxctl create deployment --help' for usage"},
				)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			createDeploymentCmdImage := args[0]
			createDeploymentCmdApp, _ := cmd.Flags().GetString("app")
			createDeploymentCmdSelector, _ := cmd.Flags().GetString("selector")

			if createDeploymentCmdApp == "" {
				return errors.E(
					errors.Invalid,
					"required argument APP not found",
					[]string{"run 'gxctl create deployment --help' for usage"},
				)
			}
			if createDeploymentCmdSelector == "" {
				return errors.E(
					errors.Invalid,
					"required argument SELECTOR not found",
					[]string{"run 'gxctl create deployment --help' for usage"},
				)
			}

			matchByLabels := make(map[string]string)
			var err error
			if createDeploymentCmdSelector != "" {
				matchByLabels, err = api.ParseMetadataMap(createDeploymentCmdSelector)
				if err != nil {
					return err
				}
			}

			image := createDeploymentCmdImage
			name, err := api.GetDockerImageName(image)
			if err != nil {
				return err
			}

			app := createDeploymentCmdApp

			d := api.CreateDeployment{Spec: &deploymentsApi.DeviceDeploymentSpec{
				App: app,
				Selector: deploymentsApi.Selector{
					MatchByLabels: matchByLabels,
				},
				Template: deploymentsApi.PodTemplate{
					Spec: podsApi.PodConfig{
						Containers: []podsApi.Container{
							{
								Name:  name,
								Image: image,
							},
						},
					},
				},
			}}

			message, err := createResource(client, d)
			if err != nil {
				return err
			}

			fmt.Println(message)
			return nil
		},
	}

	createDeploymentCmd.Flags().StringP("app", "a", "", "App for the deployment")
	createDeploymentCmd.Flags().StringP("selector", "s", "", "A space seperated list of labels to match a device eg. gridx.de/channel=stable gridx.de/area=west-1")

	createDeploymentCmd.SetHelpTemplate(template.HelpTemplate())
	createDeploymentCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(createDeploymentCmd)

	return &CreateDeployment{
		Command: createDeploymentCmd,
	}
}
