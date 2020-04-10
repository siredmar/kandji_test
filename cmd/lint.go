package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/internal/lint"
	"github.com/grid-x/gxctl/internal/lint/result"
	"github.com/grid-x/gxctl/pkg/api"
	"github.com/grid-x/gxctl/pkg/template"
)

type Lint struct {
	Command *cobra.Command
}

var fileName string

func NewLint(parent *cobra.Command) *Lint {
	var lintCmd = &cobra.Command{
		Use:                   "lint [OPTIONS]",
		DisableFlagsInUseLine: true,
		Short:                 "Lint resources",
		Long:                  `TODO`,
		Example:               "# Lint resource file \n  gxctl lint -f deployment.json",
		RunE: func(cmd *cobra.Command, args []string) error {
			lintCmdFilename, _ := cmd.Flags().GetString("filename")

			if lintCmdFilename == "" {
				cmd.Usage()
				return nil
			}

			contents, err := api.GetFilesContentsToProcess(lintCmdFilename)
			if err != nil {
				return err
			}

			results := []result.Result{}
			for _, c := range contents {
				res, err := execLint(c)
				if err != nil {
					return err
				}
				results = append(results, res...)
			}

			for _, res := range results {
				fmt.Println(res)
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

func execLint(content []byte) ([]result.Result, error) {
	res, _, err := checkResourceFile(content, true)
	if err != nil {
		return nil, err
	}

	return lint.Lint(res)
}
