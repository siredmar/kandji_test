package cmd

import (
	"github.com/spf13/cobra"

	template "github.com/grid-x/gxctl/pkg/template"
)

type ConfigCreate struct {
	Command *cobra.Command
}

func NewConfigCreate(parent *cobra.Command) *ConfigCreate {
	var configCreateCmd = &cobra.Command{
		Use:                   "create [OPTIONS]",
		DisableFlagsInUseLine: true,
		Short:                 "create different device configurations",
		Long:                  `TODO`,
	}

	configCreateCmd.SetHelpTemplate(template.HelpTemplate())
	configCreateCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(configCreateCmd)

	return &ConfigCreate{
		Command: configCreateCmd,
	}
}
