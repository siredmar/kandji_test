package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	appsv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/apps/v1beta1"
	corev1beta1 "github.com/grid-x/ds-k8s/pkg/apis/core/v1beta1"
	api "github.com/grid-x/gxctl/pkg/api"
	client "github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	template "github.com/grid-x/gxctl/pkg/template"
)

type CreateDeployment struct {
	Command *cobra.Command
}

func NewCreateDeployment(parent *cobra.Command) *CreateDeployment {
	var createDeploymentCmd = &cobra.Command{
		Use:                   "deployment IMAGE --app=APP --labels=LABELS [OPTIONS]",
		Short:                 "Creates an deployment",
		DisableFlagsInUseLine: true,
		Long:                  `TODO`,
		Aliases:               []string{"deployments", "deploy"},
		Example:               "# Create an nginx deployment for app test \n  gxctl create deployment nginx:1.15.8 -a test -l gridx.de/channel:stable",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.MissingParameter("IMAGE", "gxctl create deployment -h")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			createDeploymentCmdImage := args[0]
			createDeploymentCmdApp, _ := cmd.Flags().GetString("app")
			createDeploymentCmdLabels, _ := cmd.Flags().GetString("labels")

			if createDeploymentCmdApp == "" {
				return errors.MissingParameter("APP", "gxctl create deployment -h")
			}
			if createDeploymentCmdLabels == "" {
				return errors.MissingParameter("LABELS", "gxctl create deployment -h")
			}

			matchByLabels := make(map[string]string)
			if createDeploymentCmdLabels != "" {
				labels := strings.Split(createDeploymentCmdLabels, ",")
				for _, pair := range labels {
					z := strings.Split(pair, ":")
					matchByLabels[z[0]] = z[1]
				}
			}
			image := createDeploymentCmdImage
			name := strings.Split(createDeploymentCmdImage, ":")
			app := createDeploymentCmdApp

			client := client.NewAPIClient()
			d := api.CreateDeployment{Spec: &appsv1beta1.DeviceDeploymentSpec{
				App: app,
				Selector: appsv1beta1.Selector{
					MatchByLabels: matchByLabels,
				},
				Template: appsv1beta1.PodTemplate{
					Spec: corev1beta1.PodConfig{
						Containers: []corev1beta1.Container{
							{
								Name:  name[0],
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
	createDeploymentCmd.Flags().StringP("labels", "l", "", "A comma seperated list of labels eg. gridx.de/channel:stable,gridx.de/area:west-1")

	createDeploymentCmd.SetHelpTemplate(template.HelpTemplate())
	createDeploymentCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(createDeploymentCmd)

	return &CreateDeployment{
		Command: createDeploymentCmd,
	}
}
