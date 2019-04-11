package cmd

import (
	"github.com/spf13/cobra"

	"github.com/grid-x/gxctl/pkg/template"
)

type Label struct {
	Command *cobra.Command
}

func NewLabel(parent *cobra.Command) *Label {
	var labelCmd = &cobra.Command{
		Use:                   "label [OPTIONS]",
		DisableFlagsInUseLine: true,
		Short:                 "Label different resources",
		Long:                  `TODO`,
	}

	labelCmd.SetHelpTemplate(template.HelpTemplate())
	labelCmd.SetUsageTemplate(template.UsageTemplate())
	parent.AddCommand(labelCmd)

	return &Label{
		Command: labelCmd,
	}
}
