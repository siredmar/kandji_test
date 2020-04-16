package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/internal/lint"
	"github.com/grid-x/gxctl/internal/lint/context"
	"github.com/grid-x/gxctl/internal/lint/result"
	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/client"
	"github.com/grid-x/gxctl/pkg/errors"
	"github.com/grid-x/gxctl/pkg/template"
)

type Lint struct {
	Command *cobra.Command
}

func NewLint(parent *cobra.Command, client *client.APIClient) *Lint {
	var lintCmd = &cobra.Command{
		Use:                   "lint [OPTIONS]",
		DisableFlagsInUseLine: true,
		Short:                 "Lint resources by evaluating both desired (local files) and current (remote) state",
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
			toLint := make(map[string]interface{})

			var resources []interface{}
			for fileName, x := range fsContents {
				res, _, err := checkResourceFile(x, false) // true will return (Known after apply) for unset IDs
				if err != nil {
					fmt.Printf("SKIP %v: %v\n", fileName, err)
					continue
				}
				toLint[fileName] = res
				resources = append(resources, res)
			}

			ctx, err := context.New(client, resources)
			if err != nil {
				return err
			}

			results := []result.Result{}
			for fileName, resource := range toLint {
				result, _ := lint.Lint(ctx, fileName, resource)
				results = append(results, result...)
			}

			var total, pass, skip int
			for _, result := range results {
				if result.Skip {
					skip++
					continue
				}
				total++
				if result.Pass {
					pass++
				}
			}

			fmt.Printf("%v / %v passed (%v skipped)\n", pass, total, skip)
			for _, result := range results {
				fmt.Println(result)
			}

			if pass != total {
				return errors.E(
					errors.Validation,
					"Some linting rules did not pass",
				)
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
