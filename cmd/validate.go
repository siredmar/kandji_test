package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/template"
)

type Validate struct {
	Command *cobra.Command
}

func NewValidate(parent *cobra.Command, client *client.APIClient) *Validate {
	var validateCmd = &cobra.Command{
		Use:                   "validate [OPTIONS]",
		Short:                 "validate resources",
		DisableFlagsInUseLine: true,
		Long:                  `TODO`,
		Example:               "# Validate a resource from file \n  gxctl validate -f deployment.json",
		RunE: func(cmd *cobra.Command, args []string) error {
			validateCmdFilename, _ := cmd.Flags().GetString("filename")

			if validateCmdFilename == "" {
				cmd.Usage()
				return nil
			}

			contents, err := api.GetFilesContentsToProcess(validateCmdFilename)
			if err != nil {
				return err
			}

			for n, c := range contents {
				if err := validate(n, c, client); err != nil {
					return err
				}
			}

			return nil
		},
	}

	validateCmd.Flags().StringP("filename", "f", "", "Filename or directory to file to use to create the resource")

	validateCmd.SetHelpTemplate(template.HelpTemplate())
	validateCmd.SetUsageTemplate(template.UsageTemplate())

	parent.AddCommand(validateCmd)

	return &Validate{
		Command: validateCmd,
	}
}

func validate(filename string, content []byte, client *client.APIClient) error {
	_, _, err := checkResourceFile(content)
	if err != nil {
		return err
	}

	fmt.Println("Valid Resource")
	return nil
}
