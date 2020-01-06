package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	"github.com/grid-x/gxctl/pkg/template"
)

type Create struct {
	Command *cobra.Command
}

func NewCreate(parent *cobra.Command, client *client.APIClient) *Create {
	var createCmd = &cobra.Command{
		Use:                   "create [OPTIONS]",
		Short:                 "Create different resources",
		DisableFlagsInUseLine: true,
		Long:                  `TODO`,
		Example:               "# Create a deployment resource from file \n  gxctl create -f deployment.json",
		RunE: func(cmd *cobra.Command, args []string) error {
			createCmdFilename, _ := cmd.Flags().GetString("filename")

			if createCmdFilename == "" {
				cmd.Usage()
				return nil
			}

			contents, err := api.GetFilesContentsToProcess(createCmdFilename)
			if err != nil {
				return err
			}

			for _, c := range contents {
				if err := create(c, client); err != nil {
					return err
				}
			}

			return nil
		},
	}

	createCmd.Flags().StringP("filename", "f", "", "Filename or directory to file to use to create the resource")

	createCmd.SetHelpTemplate(template.HelpTemplate())
	createCmd.SetUsageTemplate(template.UsageTemplate())

	parent.AddCommand(createCmd)

	return &Create{
		Command: createCmd,
	}
}

func create(content []byte, client *client.APIClient) error {
	res, _, err := checkResourceFile(content, false)
	if err != nil {
		return err
	}

	message, err := createResource(client, res)
	if err != nil {
		return err
	}

	fmt.Println(message)
	return nil
}

func createResource(client *client.APIClient, v interface{}) (string, error) {
	switch v := v.(type) {
	case api.Device:
		response, err := client.PostRequest(api.DevicesEndpoint, v)
		if err != nil {
			return "", err
		}

		device, err := api.NewDevice(response)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Device %s created successfully", device.Metadata.ID), nil
	case api.Deployment:
		response, err := client.PostRequest(api.DeploymentsEndpoint, v)
		if err != nil {
			return "", err
		}

		deployment, err := api.NewDeployment(response)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Deployment %s created successfully", deployment.Metadata.ID), nil
	case api.Application:
		response, err := client.PostRequest(api.ApplicationsEndpoint, v)
		if err != nil {
			return "", err
		}

		application, err := api.NewApplication(response)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Application %s created successfully", application.Name), nil
	case api.MaintenanceTask:
		response, err := client.PostRequest(api.MaintenanceEndpoint, v)
		if err != nil {
			return "", err
		}

		task, err := api.NewMaintenanceTask(response)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Maintenance task %s created successfully", task.Metadata.ID), nil
	case api.DockerConfig:
		response, err := client.PostRequest(api.DockerConfigsEndpoint, v)
		if err != nil {
			return "", err
		}

		config, err := api.NewDockerConfig(response)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Docker config %s created successfully", config.Metadata.ID), nil
	case api.CleanupConfig:
		response, err := client.PostRequest(api.CleanupConfigsEndpoint, v)
		if err != nil {
			return "", err
		}

		config, err := api.NewCleanupConfig(response)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Cleanup config %s created successfully", config.Metadata.ID), nil
	default:
		s := fmt.Sprintf("Creating resource of type %s.", v)
		return "", errors.NotImplementedError(s)
	}
}
