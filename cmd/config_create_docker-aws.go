package cmd

import (
	"fmt"
	"strings"

	appsv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/apps/v1beta1"
	configv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/config/v1beta1"
	"github.com/spf13/cobra"

	api "github.com/grid-x/gxctl/pkg/api"
	client "github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	template "github.com/grid-x/gxctl/pkg/template"
)

type ConfigCreateDockerAWS struct {
	Command *cobra.Command
}

func NewConfigCreateDockerAWS(parent *cobra.Command) *ConfigCreateDockerAWS {
	var configCreateDockerAWSCmd = &cobra.Command{
		Use:                   "docker-aws REGISTRY --access-key=ACCESS_KEY --secret-access-key=SECRET_ACCESS_KEY --region=REGION --selector=SELECTOR [OPTIONS]",
		DisableFlagsInUseLine: true,
		Short:                 "Creates a aws docker config",
		Long:                  `TODO`,
		Example:               "# Create a docker config \n  gxctl config create docker-aws https://485611583707.dkr.ecr.eu-central-1.amazonaws.com --access-key=FOO --secret-access-key=BAR --region=eu-central-1 --selector gridx.de/channel=stable",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.MissingParameter("REGISTRY", "gxctl config create docker-aws -h")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			configCreateDockerAWSCmdRegistry := args[0]
			configCreateDockerAWSCmdAccessKey, _ := cmd.Flags().GetString("access-key")
			configCreateDockerAWSCmdSecretAccessKey, _ := cmd.Flags().GetString("secret-access-key")
			configCreateDockerAWSCmdRegion, _ := cmd.Flags().GetString("region")
			configCreateDockerAWSCmdSelector, _ := cmd.Flags().GetString("selector")

			if configCreateDockerAWSCmdAccessKey == "" {
				return errors.MissingParameter("ACCESS_KEY", "gxctl config create docker-aws -h")
			}
			if configCreateDockerAWSCmdSecretAccessKey == "" {
				return errors.MissingParameter("SECRET_ACCESS_KEY", "gxctl config create docker-aws -h")
			}
			if configCreateDockerAWSCmdRegion == "" {
				return errors.MissingParameter("REGION", "gxctl config create docker-aws -h")
			}

			matchByLabels := make(map[string]string)
			if configCreateDockerAWSCmdSelector != "" {
				labels := strings.Split(configCreateDockerAWSCmdSelector, " ")
				for _, pair := range labels {
					z := strings.Split(pair, "=")
					matchByLabels[z[0]] = z[1]
				}
			}

			client := client.NewAPIClient()
			d := api.CreateDockerConfig{
				Spec: &configv1beta1.DockerConfigSpec{
					Registry: configCreateDockerAWSCmdRegistry,
					Selector: appsv1beta1.Selector{
						MatchByLabels: matchByLabels,
					},
					Credentials: configv1beta1.DockerConfigCredentails{
						AWS: &configv1beta1.AWSCredentialProvider{
							AccessKey:       configCreateDockerAWSCmdAccessKey,
							SecretAccessKey: configCreateDockerAWSCmdSecretAccessKey,
							Region:          configCreateDockerAWSCmdRegion,
						},
					},
				},
			}

			message, err := createResource(client, d)
			if err != nil {
				return err
			}

			fmt.Println(message)
			return nil
		},
	}

	configCreateDockerAWSCmd.Flags().StringP("access-key", "a", "", "AWS ACCESS_KEY")
	configCreateDockerAWSCmd.Flags().StringP("secret-access-key", "k", "", "AWS SECRET_ACCESS_KEY")
	configCreateDockerAWSCmd.Flags().StringP("region", "r", "", "AWS Region")
	configCreateDockerAWSCmd.Flags().StringP("selector", "s", "", "A space seperated list of labels to match a device eg. gridx.de/channel=stable gridx.de/area=west-1")

	configCreateDockerAWSCmd.SetHelpTemplate(template.HelpTemplate())
	configCreateDockerAWSCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(configCreateDockerAWSCmd)

	return &ConfigCreateDockerAWS{
		Command: configCreateDockerAWSCmd,
	}
}
