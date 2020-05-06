package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
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

			resources := make(map[string]interface{}, len(contents))

			for _, c := range contents {
				res, resID, err := checkResourceFile(c, false)
				if err != nil {
					return err
				}
				resources[resID] = res
			}

			resourcesSorted := sortByKind(resources)

			for _, r := range resourcesSorted {
				if err := create(r.ID, r.Res, client); err != nil {
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

func create(resID string, res interface{}, client *client.APIClient) error {
	message, err := createResource(resID, res, client)
	if err != nil {
		return err
	}

	fmt.Println(message)
	return nil
}

func createResource(resID string, res interface{}, client *client.APIClient) (string, error) {
	switch res := res.(type) {
	case api.Device:
		response, err := client.PostRequest(api.DevicesEndpoint, res)
		if err != nil {
			return "", err
		}
		if _, err := api.NewDevice(response); err != nil {
			return "", err
		}
		return fmt.Sprintf("Device %s created successfully", resID), nil

	case api.Deployment:
		response, err := client.PostRequest(api.DeploymentsEndpoint, res)
		if err != nil {
			return "", err
		}
		if _, err := api.NewDeployment(response); err != nil {
			return "", err
		}
		return fmt.Sprintf("Deployment %s created successfully", resID), nil

	case api.Application:
		response, err := client.PostRequest(api.ApplicationsEndpoint, res)
		if err != nil {
			return "", err
		}
		if _, err := api.NewApplication(response); err != nil {
			return "", err
		}
		return fmt.Sprintf("Application %s created successfully", resID), nil

	case api.MaintenanceTask:
		response, err := client.PostRequest(api.MaintenanceEndpoint, res)
		if err != nil {
			return "", err
		}
		if _, err := api.NewMaintenanceTask(response); err != nil {
			return "", err
		}
		return fmt.Sprintf("Maintenance task %s created successfully", resID), nil

	case api.DockerConfig:
		response, err := client.PostRequest(api.DockerConfigsEndpoint, res)
		if err != nil {
			return "", err
		}
		if _, err := api.NewDockerConfig(response); err != nil {
			return "", err
		}
		return fmt.Sprintf("Docker config %s created successfully", resID), nil

	case api.CleanupConfig:
		response, err := client.PostRequest(api.CleanupConfigsEndpoint, res)
		if err != nil {
			return "", err
		}
		if _, err := api.NewCleanupConfig(response); err != nil {
			return "", err
		}
		return fmt.Sprintf("Cleanup config %s created successfully", resID), nil

	default:
		return "", errors.E(
			errors.NotImplemented,
			"Unsupported type",
		)
	}
}
