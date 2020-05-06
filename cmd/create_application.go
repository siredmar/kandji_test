package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
	"github.com/grid-x/gxctl/pkg/template"
)

type CreateApplication struct {
	Command *cobra.Command
}

func NewCreateApplication(parent *cobra.Command, client *client.APIClient) *CreateApplication {
	var createApplicationCmd = &cobra.Command{
		Use:                   "application NAME [OPTIONS]",
		Short:                 "Creates an application",
		DisableFlagsInUseLine: true,
		Aliases:               []string{"applications", "app", "apps"},
		Example:               "# Create an application with name test \n  gxctl create application test",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.E(
					errors.Invalid,
					"required argument NAME not found",
					[]string{"run 'gxctl create application --help' for usage"},
				)
			}
			return nil
		},
		Long: `TODO`,
		RunE: func(cmd *cobra.Command, args []string) error {
			createApplicationCmdName := args[0]
			d := api.Application{}

			if createApplicationCmdName != "" {
				d.Name = createApplicationCmdName
			}

			message, err := createResource(createApplicationCmdName, d, client)
			if err != nil {
				return err
			}

			fmt.Println(message)
			return nil
		},
	}

	parent.AddCommand(createApplicationCmd)

	createApplicationCmd.SetHelpTemplate(template.HelpTemplate())
	createApplicationCmd.SetUsageTemplate(template.UsageTemplate())

	return &CreateApplication{
		Command: createApplicationCmd,
	}
}
