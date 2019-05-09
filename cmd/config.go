package cmd

import (
	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/pkg/template"
)

type Config struct {
	Command *cobra.Command
}

func NewConfig(parent *cobra.Command) *Config {
	var configCmd = &cobra.Command{
		Use:                   "config [OPTIONS]",
		DisableFlagsInUseLine: true,
		Short:                 "Configure devices",
		Long:                  `TODO`,
	}

	configCmd.SetHelpTemplate(template.HelpTemplate())
	configCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(configCmd)

	return &Config{
		Command: configCmd,
	}
}
