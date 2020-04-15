package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/internal/lint"
	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/result"
	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/template"
)

type Lint struct {
	Command *cobra.Command
}

func NewLint(parent *cobra.Command, client *client.APIClient) *Lint {
	var lintCmd = &cobra.Command{
		Use:                   "lint [OPTIONS]",
		DisableFlagsInUseLine: true,
		Short:                 "Lint resources by evaluating local files and remote state",
		Long:                  `TODO`,
		Example:               "# Lint resource file \n  gxctl lint -f deployment.json",
		RunE: func(cmd *cobra.Command, args []string) error {
			lintCmdFilename, _ := cmd.Flags().GetString("filename")

			if lintCmdFilename == "" {
				cmd.Usage()
				return nil
			}

			fsContents, err := api.GetFilesContentsToProcess(lintCmdFilename)
			if err != nil {
				return err
			}

			var resources []interface{}
			for _, x := range fsContents {
				res, _, err := checkResourceFile(x, true)
				if err != nil {
					return err
				}
				resources = append(resources, res)
			}

			ctx, err := context.New(client, resources)
			if err != nil {
				return err
			}

			results := []result.Result{}
			for fileName, c := range fsContents {
				result, err := execLint(ctx, fileName, c)
				if err != nil {
					return err
				}
				results = append(results, result...)
			}

			for _, result := range results {
				fmt.Println(result)
			}

			return nil
		},
	}

	lintCmd.Flags().StringP("filename", "f", "", "filename or directory of files to lint")

	lintCmd.SetHelpTemplate(template.HelpTemplate())
	lintCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(lintCmd)

	return &Lint{
		Command: lintCmd,
	}
}

func execLint(ctx *context.Context, fileName string, content []byte) ([]result.Result, error) {
	res, _, err := checkResourceFile(content, true)
	if err != nil {
		return nil, err
	}

	return lint.Lint(ctx, fileName, res)
}
