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

			// read raw file contents into fileName -> bytes map
			fsContents, err := api.GetFilesContentsToProcess(lintCmdFilename)
			if err != nil {
				return err
			}

			// parse file contents into fileName -> resource map
			fsResources := make(map[string]interface{})
			var fsNames []string
			for fileName, x := range fsContents {
				res, _, err := checkResourceFile(x, false) // true will return (Known after apply) for unset IDs
				if err != nil {
					fmt.Printf("SKIP %v: %v\n", fileName, err)
					continue
				}
				fsNames = append(fsNames, fileName)
				fsResources[fileName] = res
			}

			// init context with current state
			ctx, err := context.New(client)
			if err != nil {
				return err
			}

			// build permutations:
			// for each file, we need a context which includes every other file but not itself
			jobs := make(map[string][]string)
			for i, x := range fsNames {
				tmp := make([]string, len(fsNames))
				copy(tmp, fsNames)
				jobs[x] = append(tmp[:i], tmp[i+1:]...)
			}

			results := []result.Result{}
			for fileName, context := range jobs {
				desired := []interface{}{}

				for _, ctxRes := range context {
					desired = append(desired, fsResources[ctxRes])
				}
				ctx.SetDesired(desired)

				result, _ := lint.Lint(ctx, fileName, fsResources[fileName])
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
