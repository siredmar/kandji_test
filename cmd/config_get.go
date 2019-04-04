package cmd

import (
	"github.com/spf13/cobra"

	template "github.com/grid-x/gxctl/pkg/template"
)

type ConfigGet struct {
	Command *cobra.Command
}

func NewConfigGet(parent *cobra.Command) *ConfigGet {
	var configGetCmd = &cobra.Command{
		Use:                   "get [OPTIONS]",
		DisableFlagsInUseLine: true,
		Short:                 "get different device configurations",
		Long:                  `TODO`,
	}

	configGetCmd.PersistentFlags().StringP("output", "o", "", "print the result in a different format. Currently supported json/wide/yaml")
	configGetCmd.SetHelpTemplate(template.HelpTemplate())
	configGetCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(configGetCmd)

	return &ConfigGet{
		Command: configGetCmd,
	}
}
