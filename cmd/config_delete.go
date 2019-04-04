package cmd

import (
	"github.com/spf13/cobra"

	template "github.com/grid-x/gxctl/pkg/template"
)

type ConfigDelete struct {
	Command *cobra.Command
}

func NewConfigDelete(parent *cobra.Command) *ConfigDelete {
	var configDeleteCmd = &cobra.Command{
		Use:                   "delete [OPTIONS]",
		DisableFlagsInUseLine: true,
		Short:                 "delete different device configurations",
		Long:                  `TODO`,
	}

	configDeleteCmd.SetHelpTemplate(template.HelpTemplate())
	configDeleteCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(configDeleteCmd)

	return &ConfigDelete{
		Command: configDeleteCmd,
	}
}
