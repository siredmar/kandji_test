package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	errors "github.com/grid-x/gxctl/pkg/error"
	"github.com/grid-x/gxctl/pkg/template"
)

type Apply struct {
	Command *cobra.Command
}

func NewApply(parent *cobra.Command, client *client.APIClient) *Apply {
	var applyCmd = &cobra.Command{
		Use:                   "apply [OPTIONS]",
		Short:                 "apply resources",
		DisableFlagsInUseLine: true,
		Long:                  `TODO`,
		Example:               "# Apply a resource from file \n  gxctl apply -f deployment.json",
		RunE: func(cmd *cobra.Command, args []string) error {
			applyCmdFilename, _ := cmd.Flags().GetString("filename")

			if applyCmdFilename == "" {
				cmd.Usage()
				return nil
			}

			contents, err := api.GetFilesContentsToProcess(applyCmdFilename)
			if err != nil {
				return err
			}

			for _, c := range contents {
				if err := apply(c, client); err != nil {
					return err
				}
			}

			return nil
		},
	}

	applyCmd.Flags().StringP("filename", "f", "", "Filename or directory to file to use to create the resource")

	applyCmd.SetHelpTemplate(template.HelpTemplate())
	applyCmd.SetUsageTemplate(template.UsageTemplate())

	parent.AddCommand(applyCmd)

	return &Apply{
		Command: applyCmd,
	}
}

func apply(content []byte, client *client.APIClient) error {
	res, resId, err := checkResourceFile(content, false)
	if err != nil {
		return err
	}

	if resId == "" {
		// Create
		return create(content, client)
	}

	// Update or Create
	found := false
	switch v := res.(type) {
	case api.Device:
		_, err := getDeviceById(client, v.Metadata.ID, nil)
		if err == nil {
			found = true
		}
	case api.Deployment:
		_, err := getDeploymentById(client, v.Metadata.ID, nil)
		if err == nil {
			found = true
		}
	case api.DockerConfig:
		_, err := getDockerConfigById(client, v.Metadata.ID, nil)
		if err == nil {
			found = true
		}
	case api.CleanupConfig:
		_, err := getCleanupConfigById(client, v.Metadata.ID, nil)
		if err == nil {
			found = true
		}
	default:
		s := fmt.Sprintf("Updating resource of type %s.", v)
		return errors.NotImplementedError(s)
	}

	if !found {
		// Create
		return create(content, client)
	}

	// Update
	return update(content, client)
}
