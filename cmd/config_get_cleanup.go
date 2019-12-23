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

type ConfigGetCleanup struct {
	Command *cobra.Command
}

func NewConfigGetCleanup(parent *cobra.Command, client *client.APIClient, printer *print.Printer) *ConfigGetCleanup {
	var configGetCleanupCmd = &cobra.Command{
		Use:                   "cleanup [OPTIONS]",
		DisableFlagsInUseLine: true,
		Short:                 "retrieves cleanup configs",
		Example:               "# Get all cleanup configs \n  gxctl config get cleanup\n\n  # Get information about an cleanup config with abbreviation c72 \n  gxctl config get cleanup c72",
		Long:                  `Prints a list of all cleanup configs you have access to`,
		RunE: func(cmd *cobra.Command, args []string) error {
			getCmdOutputType, _ := cmd.Flags().GetString("output")

			printerConfig := print.Printconfig{
				OutputFormat: getCmdOutputType,
			}

			if len(args) > 0 {
				//Get multiple cleanupConfigs
				//Lookup all existing cleanupConfigs to validate ids and autocomplete them if necessary
				cleanupconfigs, err := getCleanupConfigs(client)
				if err != nil {
					return err
				}
				cleanupConfigIDs := cleanupconfigs.GetIds()

				for _, a := range args {
					config, err := getCleanupConfigById(client, a, cleanupConfigIDs)
					if err != nil {
						return err
					}

					if err := printer.Print(config, printerConfig); err != nil {
						return err
					}
				}
			} else {
				//List all devices
				cleanupConfigs, err := getCleanupConfigs(client)
				if err != nil {
					return err
				}

				if err := printer.Print(cleanupConfigs, printerConfig); err != nil {
					return err
				}
			}
			return nil
		},
	}

	configGetCleanupCmd.SetHelpTemplate(template.HelpTemplate())
	configGetCleanupCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(configGetCleanupCmd)

	return &ConfigGetCleanup{
		Command: configGetCleanupCmd,
	}
}

func getCleanupConfigs(client *client.APIClient) (api.CleanupConfigs, error) {
	response, err := client.GetRequest(api.CleanupConfigsEndpoint)
	if err != nil {
		return api.CleanupConfigs{}, err
	}

	cleanupConfigList, err := api.NewCleanupConfigs(response)
	if err != nil {
		return cleanupConfigList, err
	}

	if cleanupConfigList.IsEmpty() {
		return cleanupConfigList, errors.ListNotFoundError("cleanupConfigs")
	}

	return cleanupConfigList, nil
}

func getCleanupConfigById(client *client.APIClient, id string, cleanupConfigIDs []string) (api.CleanupConfig, error) {
	configID, err := api.LookupID(id, cleanupConfigIDs)
	if err != nil {
		return api.CleanupConfig{}, err
	}

	endpoint := fmt.Sprintf("%s/%s", api.CleanupConfigsEndpoint, configID)
	response, err := client.GetRequest(endpoint)
	if err != nil {
		return api.CleanupConfig{}, err
	}

	config, err := api.NewCleanupConfig(response)
	if err != nil {
		return config, err
	}

	return config, nil
}
