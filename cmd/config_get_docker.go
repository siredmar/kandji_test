package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	"github.com/grid-x/gxctl/pkg/printer"
	"github.com/grid-x/gxctl/pkg/template"
)

type ConfigGetDocker struct {
	Command *cobra.Command
}

func NewConfigGetDocker(parent *cobra.Command, client *client.APIClient, printer *printer.Printer) *ConfigGetDocker {
	var configGetDockerCmd = &cobra.Command{
		Use:                   "docker [OPTIONS]",
		DisableFlagsInUseLine: true,
		Short:                 "retrieves docker configs",
		Example:               "# Get all docker configs \n  gxctl config get docker\n\n  # Get information about an docker config with abbreviation c72 \n  gxctl config get docker c72",
		Long:                  `Prints a list of all docker configs you have access to`,
		RunE: func(cmd *cobra.Command, args []string) error {
			getCmdOutputType, _ := cmd.Flags().GetString("output")

			if len(args) > 0 {
				//Get multiple dockerConfigs
				//Lookup all existing dockerConfigs to validate ids and autocomplete them if necessary
				dockerconfigs, err := getDockerConfigs(client)
				if err != nil {
					return err
				}
				dockerConfigIDs := dockerconfigs.GetIds()

				for _, a := range args {
					config, err := getDockerConfigById(client, a, dockerConfigIDs)
					if err != nil {
						return err
					}

					if err := printer.Print(config, getCmdOutputType); err != nil {
						return err
					}
				}
			} else {
				//List all devices
				dockerConfigs, err := getDockerConfigs(client)
				if err != nil {
					return err
				}

				if err := printer.Print(dockerConfigs, getCmdOutputType); err != nil {
					return err
				}
			}
			return nil
		},
	}

	configGetDockerCmd.SetHelpTemplate(template.HelpTemplate())
	configGetDockerCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(configGetDockerCmd)

	return &ConfigGetDocker{
		Command: configGetDockerCmd,
	}
}

func getDockerConfigs(client *client.APIClient) (api.DockerConfigs, error) {
	response, err := client.GetRequest(api.DockerConfigsEndpoint)
	if err != nil {
		return api.DockerConfigs{}, err
	}

	dockerConfigList, err := api.NewDockerConfigs(response)
	if err != nil {
		return dockerConfigList, err
	}

	if dockerConfigList.IsEmpty() {
		return dockerConfigList, errors.ListNotFoundError("dockerConfigs")
	}

	return dockerConfigList, nil
}

func getDockerConfigById(client *client.APIClient, id string, dockerConfigIDs []string) (api.DockerConfig, error) {
	configID, err := api.LookupID(id, dockerConfigIDs)
	if err != nil {
		return api.DockerConfig{}, err
	}

	endpoint := fmt.Sprintf("%s/%s", api.DevicesEndpoint, configID)
	response, err := client.GetRequest(endpoint)
	if err != nil {
		return api.DockerConfig{}, err
	}

	config, err := api.NewDockerConfig(response)
	if err != nil {
		return config, err
	}

	return config, nil
}
