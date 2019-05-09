package cmd

import (
	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/pkg/template"
)

type Get struct {
	Command *cobra.Command
}

func NewGet(parent *cobra.Command) *Get {
	var getCmd = &cobra.Command{
		Use:                   "get [OPTIONS]",
		DisableFlagsInUseLine: true,
		Short:                 "Get different resources",
		Long:                  `TODO`,
	}

	getCmd.SetHelpTemplate(template.HelpTemplate())
	getCmd.SetUsageTemplate(template.UsageTemplate())
	getCmd.PersistentFlags().StringP("output", "o", "", "print the result in a different format. Currently supported json/wide/yaml")
	parent.AddCommand(getCmd)

	return &Get{
		Command: getCmd,
	}
}
