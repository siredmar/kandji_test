package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
	"github.com/grid-x/gxctl/pkg/template"
)

type Delete struct {
	Command *cobra.Command
}

func NewDelete(parent *cobra.Command, client *client.APIClient) *Delete {
	var deleteCmd = &cobra.Command{
		Use:                   "delete [OPTIONS]",
		Short:                 "Delete one or multiple resources from a file or directory",
		Long:                  `TODO`,
		DisableFlagsInUseLine: true,
		Example:               "# Delete a resource from file\n  gxctl delete -f deployment.json",
		RunE: func(cmd *cobra.Command, args []string) error {
			deleteCmdFilename, _ := cmd.Flags().GetString("filename")

			if deleteCmdFilename == "" {
				cmd.Usage()
				return nil
			}

			contents, err := api.GetFilesContentsToProcess(deleteCmdFilename)
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

			for resID, res := range resources {
				if err := deleteResource(resID, res, client); err != nil {
					return err
				}
			}

			return nil
		},
	}

	deleteCmd.Flags().StringP("filename", "f", "", "Filename or directory of files")
	deleteCmd.SetHelpTemplate(template.HelpTemplate())
	deleteCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(deleteCmd)

	return &Delete{
		Command: deleteCmd,
	}
}

func deleteResource(resID string, res interface{}, client *client.APIClient) error {
	fmt.Printf("deleting resource %s… ", resID)

	switch res.(type) {
	case api.Application:
		if _, err := client.DeleteRequest(api.ApplicationsEndpoint, resID); err != nil {
			return err
		}
		break

	case api.Device:
		if _, err := client.DeleteRequest(api.DevicesEndpoint, resID); err != nil {
			return err
		}
		break

	case api.Deployment:
		if _, err := client.DeleteRequest(api.DockerConfigsEndpoint, resID); err != nil {
			return err
		}
		break

	case api.DockerConfig:
		if _, err := client.DeleteRequest(api.DockerConfigsEndpoint, resID); err != nil {
			return err
		}
		break

	case api.CleanupConfig:
		if _, err := client.DeleteRequest(api.CleanupConfigsEndpoint, resID); err != nil {
			return err
		}
		break

	default:
		return errors.E(
			errors.NotImplemented,
			"Unsupported type",
		)
	}

	fmt.Print("success\n")

	return nil
}
