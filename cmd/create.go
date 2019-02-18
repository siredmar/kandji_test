package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"

	api "github.com/grid-x/gxctl/pkg/api"
	client "github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	template "github.com/grid-x/gxctl/pkg/template"
)

type Create struct {
	Command *cobra.Command
}

func NewCreate(parent *cobra.Command) *Create {
	var createCmd = &cobra.Command{
		Use:                   "create [OPTIONS]",
		Short:                 "Create different resources",
		DisableFlagsInUseLine: true,
		Long:                  `TODO`,
		Example:               "# Create a deployment resource from file \n  gxctl create -f deployment.json",
		RunE: func(cmd *cobra.Command, args []string) error {
			createCmdFilename, _ := cmd.Flags().GetString("filename")

			client := client.NewAPIClient()

			if createCmdFilename == "" {
				cmd.Usage()
				return nil
			}

			jsonFile, err := os.Open(createCmdFilename)
			if err != nil {
				return err
			}
			defer jsonFile.Close()

			bytes, err := ioutil.ReadAll(jsonFile)
			if err != nil {
				return err
			}

			res, err := checkCreateResourceFile(bytes)
			if err != nil {
				return err
			}

			message, err := createResource(client, res)
			if err != nil {
				return err
			}

			fmt.Println(message)
			return nil
		},
	}

	createCmd.Flags().StringP("filename", "f", "", "Filename to file to use to create the resource")

	createCmd.SetHelpTemplate(template.HelpTemplate())
	createCmd.SetUsageTemplate(template.UsageTemplate())

	parent.AddCommand(createCmd)

	return &Create{
		Command: createCmd,
	}
}

func checkCreateResourceFile(bytes []byte) (interface{}, error) {
	deploymentCreate, err := api.NewCreateDeployment(bytes)
	if err != nil {
		return nil, err
	}
	if !deploymentCreate.IsEmpty() && deploymentCreate.IsValid() {
		return deploymentCreate, nil
	}

	applicationCreate, err := api.NewCreateApplication(bytes)
	if err != nil {
		return nil, err
	}
	if !applicationCreate.IsEmpty() && applicationCreate.IsValid() {
		return applicationCreate, nil
	}

	//Nothing found
	return nil, errors.InvalidFormat()
}

func createResource(client *client.APIClient, v interface{}) (string, error) {
	switch v := v.(type) {
	case api.CreateDeployment:
		response, err := client.PostRequest(api.DeploymentsEndpoint, v)
		if err != nil {
			return "", err
		}

		deployment, err := api.NewDeployment(response)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Deployment %s created successfully", deployment.Metadata.ID), nil
	case api.CreateApplication:
		response, err := client.PostRequest(api.ApplicationsEndpoint, v)
		if err != nil {
			return "", err
		}

		application, err := api.NewApplication(response)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Application %s created successfully", application.Name), nil
	case api.CreateMaintenanceTask:
		response, err := client.PostRequest(api.MaintenanceEndpoint, v)
		if err != nil {
			return "", err
		}

		task, err := api.NewMaintenanceTask(response)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Maintenance task %s created successfully", task.Metadata.ID), nil
	default:
		s := fmt.Sprintf("Creating resource of type %s.", v)
		return "", errors.NotImplementedError(s)
	}
}
