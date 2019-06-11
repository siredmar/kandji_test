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

type GetApplications struct {
	Command *cobra.Command
}

func NewGetApplications(parent *cobra.Command, client *client.APIClient, printer *printer.Printer) *GetApplications {
	var getApplicationsCmd = &cobra.Command{
		Use:                   "application [NAME] [OPTIONS]",
		DisableFlagsInUseLine: true,
		Aliases:               []string{"applications", "app", "apps"},
		Short:                 "get application",
		Long:                  `Prints a list of all applications you have access to`,
		Example:               "# Get all applications \n  gxctl get applications\n\n  # Get information about an application with name test \n  gxctl get application test",
		RunE: func(cmd *cobra.Command, args []string) error {
			getCmdOutputType, _ := cmd.Flags().GetString("output")

			if len(args) > 0 {
				//Get multiple application
				for _, a := range args {
					application, err := getApplicationById(client, a)
					if err != nil {
						return err
					}

					if err := printer.Print(application, getCmdOutputType, false); err != nil {
						return err
					}
				}
			} else {
				//List all applications
				applications, err := getApplications(client)
				if err != nil {
					return err
				}

				if err := printer.Print(applications, getCmdOutputType, false); err != nil {
					return err
				}
			}
			return nil
		},
	}

	getApplicationsCmd.SetHelpTemplate(template.HelpTemplate())
	getApplicationsCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(getApplicationsCmd)

	return &GetApplications{
		Command: getApplicationsCmd,
	}
}

func getApplications(client *client.APIClient) (api.Applications, error) {
	response, err := client.GetRequest(api.ApplicationsEndpoint)
	if err != nil {
		return api.Applications{}, err
	}

	applicationList, err := api.NewApplications(response)
	if err != nil {
		return applicationList, err
	}

	if applicationList.IsEmpty() {
		return applicationList, errors.ListNotFoundError("applications")
	}

	return applicationList, nil
}

func getApplicationById(client *client.APIClient, id string) (api.Application, error) {
	endpoint := fmt.Sprintf("%s/%s", api.ApplicationsEndpoint, id)
	response, err := client.GetRequest(endpoint)
	if err != nil {
		return api.Application{}, err
	}

	application, err := api.NewApplication(response)
	if err != nil {
		return application, err
	}

	if application.IsEmpty() {
		msg := fmt.Sprintf("application \"%s\"", id)
		return application, errors.GetNotFoundError(msg)
	}

	return application, nil
}
