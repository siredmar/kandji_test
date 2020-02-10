package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
	"github.com/grid-x/gxctl/pkg/template"
)

type DeleteApplication struct {
	Command *cobra.Command
}

func NewDeleteApplication(parent *cobra.Command, client *client.APIClient) *DeleteApplication {
	var deleteApplicationCmd = &cobra.Command{
		Use:                   "application NAME [OPTIONS]",
		Aliases:               []string{"applications", "app", "apps"},
		DisableFlagsInUseLine: true,
		Short:                 "delete app",
		Long:                  `TODO`,
		Example:               "# Delete an application with name test \n  gxctl delete application test",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.E(
					errors.Invalid,
					"required argument NAME not found",
					[]string{"run 'gxctl delete application --help' for usage"},
				)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, d := range args {
				msg, err := deleteApp(client, d)
				if err != nil {
					return err
				}
				fmt.Println(msg)
			}
			return nil
		},
	}

	deleteApplicationCmd.SetHelpTemplate(template.HelpTemplate())
	deleteApplicationCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(deleteApplicationCmd)

	return &DeleteApplication{
		Command: deleteApplicationCmd,
	}
}

func deleteApp(client *client.APIClient, appID string) (string, error) {
	_, err := client.DeleteRequest(api.ApplicationsEndpoint, appID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("App %s deleted successfully", appID), nil
}
