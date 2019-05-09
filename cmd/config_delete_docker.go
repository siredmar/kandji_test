package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	"github.com/grid-x/gxctl/pkg/template"
)

type ConfigDeleteDocker struct {
	Command *cobra.Command
}

func NewConfigDeleteDocker(parent *cobra.Command, client *client.APIClient) *ConfigDeleteDocker {
	var configDeleteDockerCmd = &cobra.Command{
		Use:                   "docker ID [OPTIONS]",
		DisableFlagsInUseLine: true,
		Short:                 "delete different docker configurations",
		Long:                  `TODO`,
		Example:               "# Delete docker cofigs  \n  gxctl config docker delete 337df243-2cc9-46f4-bfeb-3c978ece4252",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.MissingParameter("ID", "gxctl config delete -h")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			//Lookup all existing dockerConfigs to validate ids and autocomplete them if necessary
			dockerConfigs, err := getDockerConfigs(client)
			if err != nil {
				return err
			}
			dockerConfigIDs := dockerConfigs.GetIds()

			for _, d := range args {
				msg, err := deleteDockerConfig(client, d, dockerConfigIDs)
				if err != nil {
					return err
				}
				fmt.Println(msg)
			}
			return nil
		},
	}

	configDeleteDockerCmd.SetHelpTemplate(template.HelpTemplate())
	configDeleteDockerCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(configDeleteDockerCmd)

	return &ConfigDeleteDocker{
		Command: configDeleteDockerCmd,
	}
}

func deleteDockerConfig(client *client.APIClient, configID string, dockerConfigIDs []string) (string, error) {
	configID, err := api.LookupID(configID, dockerConfigIDs)
	if err != nil {
		return "", err
	}

	_, err = client.DeleteRequest(api.DockerConfigsEndpoint, configID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Docker config %s deleted successfully", configID), nil
}
